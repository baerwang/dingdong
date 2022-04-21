package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

import (
	"gopkg.in/yaml.v3"
)

var (
	aid, file string
	reserve   int
	config    Info
	timeout   = 2 * time.Minute  // app timeout done
	stopTime  = time.Microsecond // stop a while
	tr        *http.Transport
)

func main() {

	handle()

	var cart, order map[string]interface{}

	m := NewConcurrentMap()

	cart = carts()

	if len(cart) != 0 {
		m.Set("cart", cart)
	}

	if len(order) != 0 {
		m.Set("order", order)
	}

	reserveMap := customizeReserve()

	notify, notifyCancel := context.WithTimeout(context.Background(), timeout)
	defer notifyCancel()

	go execute(notify, func() {
		cart = carts()
		if len(cart) != 0 {
			m.Set("cart", cart)
			return
		}
		sleeps()
	})

	go execute(notify, func() {
		storeCart, ok := m.Get("cart")
		if ok {
			order = checkOrder(aid, storeCart.(map[string]interface{}), reserveMap)
			if len(order) != 0 {
				m.Set("order", order)
				return
			}
			sleeps()
		}
	})

	ch := make(chan struct{})
	// timeout limit
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				notifyCancel()
				ch <- struct{}{}
				return
			default:
				go func() {
					storeCart, cok := m.Get("cart")
					storeOrder, ook := m.Get("order")
					if cok && ook {
						if submitOrder(aid, storeCart.(map[string]interface{}),
							storeOrder.(map[string]interface{}), reserveMap) {
							cancel()
							notifyCancel()
							ch <- struct{}{}
							return
						}
					} else {
						sleeps()
					}
				}()
			}
		}
	}()

	<-ch

}

func handle() {

	flag.StringVar(&aid, "aid", "", "address id")
	flag.StringVar(&file, "f", "", "config need info")
	flag.IntVar(&reserve, "reserve", 1, "reserve Time")
	flag.Parse()

	if file == "" {
		panic("config file not null")
	}

	readFile, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	// 将json解码
	if err = yaml.Unmarshal(readFile, &config); err != nil {
		panic(err)
	}

	// in advance can take out the address id,prevent limiting get fail
	if aid == "" {
		if aid = addressId(); aid == "" {
			panic("find user address id fail")
		}
		log.Println("获取收货人信息:" + aid)
	}

	tr = &http.Transport{MaxIdleConns: 1000, MaxIdleConnsPerHost: 100}

}

func execute(ctx context.Context, task func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			go func() {
				for i := 0; i < 3; i++ {
					task()
				}
			}()
		}
	}
}

func sleeps() {
	time.Sleep(stopTime)
}

// userInfo user info
func userInfo() url.Values {
	values := url.Values{}
	values["uid"] = []string{config.UserInfo.Uid}
	values["longitude"] = []string{config.UserInfo.Longitude}
	values["latitude"] = []string{config.UserInfo.Latitude}
	values["station_id"] = []string{config.UserInfo.StationId}
	values["city_number"] = []string{config.UserInfo.CityNumber}
	values["api_version"] = []string{config.UserInfo.ApiVersion}
	values["app_version"] = []string{config.UserInfo.AppVersion}
	values["applet_source"] = []string{config.UserInfo.AppletSource}
	values["channel"] = []string{config.UserInfo.Channel}
	values["app_client_id"] = []string{config.UserInfo.AppClientId}
	values["sharer_uid"] = []string{config.UserInfo.SharerUid}
	values["openid"] = []string{config.UserInfo.Openid}
	values["h5_source"] = []string{config.UserInfo.H5Source}
	values["s_id"] = []string{config.UserInfo.Sid}
	values["device_token"] = []string{config.UserInfo.DeviceToken}
	return values
}

// headers http header
func headers() http.Header {
	headerMap := map[string][]string{}
	headerMap["ddmc-city-number"] = []string{config.Headers.CityNumber}
	headerMap["ddmc-time"] = []string{fmt.Sprint(time.Now().Unix())}
	headerMap["ddmc-build-version"] = []string{config.Headers.BuildVersion}
	headerMap["ddmc-device-id"] = []string{config.Headers.DeviceId}
	headerMap["ddmc-station-id"] = []string{config.Headers.StationId}
	headerMap["ddmc-channel"] = []string{config.Headers.Channel}
	headerMap["ddmc-os-version"] = []string{config.Headers.OsVersion}
	headerMap["ddmc-app-client-id"] = []string{config.Headers.AppClientId}
	headerMap["cookie"] = []string{config.Headers.Cookie}
	headerMap["ddmc-ip"] = []string{config.Headers.Ip}
	headerMap["ddmc-longitude"] = []string{config.Headers.Longitude}
	headerMap["ddmc-latitude"] = []string{config.Headers.Latitude}
	headerMap["ddmc-api-version"] = []string{config.Headers.ApiVersion}
	headerMap["ddmc-uid"] = []string{config.Headers.Uid}
	headerMap["user-agent"] = []string{config.Headers.UserAgent}
	headerMap["referer"] = []string{config.Headers.Referer}
	return headerMap
}

// addressId get user address id `input choose the default`
func addressId() string {
	const _url = "https://sunquan.api.ddxq.mobi/api/v1/user/address/"

	client := http.Client{Timeout: 5 * time.Second, Transport: tr}
	request, _ := http.NewRequest(http.MethodGet, _url, nil)
	request.Header = headers()

	resp, err := client.Do(request)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	all, err := ioutil.ReadAll(resp.Body)

	info := map[string]interface{}{}
	if err = json.Unmarshal(all, &info); err != nil {
		log.Println(err)
		return ""
	}

	if !httpStatus(info, "用户") {
		return ""
	}

	rest := parseRest(info["data"])
	if len(rest) == 0 {
		return ""
	}

	item, ok := rest["valid_address"]
	if !ok {
		log.Println("address list get fail")
		return ""
	}

	marshal, _ := json.Marshal(item)
	var infos []map[string]interface{}
	if err = json.Unmarshal(marshal, &infos); err != nil {
		log.Println(err)
		return ""
	}

	for _, m := range infos {
		isDefault, ok := m["is_default"].(bool)
		if ok {
			if isDefault {
				return m["id"].(string)
			}
		}
	}

	return ""
}

func carts() map[string]interface{} {
	const _url = "https://maicai.api.ddxq.mobi/cart/index"
	client := http.Client{Timeout: 5 * time.Second, Transport: tr}
	request, _ := http.NewRequest(http.MethodGet, _url, nil)
	request.Header = headers()
	user := userInfo()
	user["is_load"] = []string{"1"}
	request.URL.RawQuery = user.Encode()

	do, err := client.Do(request)
	if err != nil {
		// log.Println(err)
		return nil
	}
	defer do.Body.Close()

	all, _ := ioutil.ReadAll(do.Body)
	info := map[string]interface{}{}
	if err = json.Unmarshal(all, &info); err != nil {
		log.Println(err)
		return nil
	}

	if !httpStatus(info, "更新购物车数据") {
		return nil
	}

	marshal, _ := json.Marshal(info["data"])
	// fmt.Printf("%s", marshal)
	if err = json.Unmarshal(marshal, &info); err != nil {
		log.Println(err)
		return nil
	}

	marshal, _ = json.Marshal(info["new_order_product_list"])
	var infos []map[string]interface{}
	if err = json.Unmarshal(marshal, &infos); err != nil {
		log.Println(err)
		return nil
	}

	if len(infos) == 0 {
		log.Println("购物车无可买的商品")
		return nil
	}

	//fmt.Printf("%s\n", marshal)
	marshal, _ = json.Marshal(infos[0]["products"])
	var products []Product
	if err = json.Unmarshal(marshal, &products); err != nil {
		log.Println(err)
		return nil
	}

	marshal, _ = json.Marshal(infos[0])
	var cart Cart
	if err = json.Unmarshal(marshal, &cart); err != nil {
		log.Println(err)
		return nil
	}

	marshal, _ = json.Marshal(info["parent_order_info"])
	if err = json.Unmarshal(marshal, &info); err != nil {
		log.Println(err)
		return nil
	}

	rst := make(map[string]interface{}, 30)

	for i := range products {
		products[i].TotalMoney = products[i].TotalPrice
		products[i].TotalOriginMoney = products[i].TotalOriginPrice
	}

	rst["products"] = products
	rst["parent_order_sign"] = info["parent_order_sign"]
	rst["total_money"] = cart.TotalMoney
	rst["goods_real_money"] = cart.GoodsRealMoney
	rst["total_count"] = cart.TotalCount
	rst["cart_count"] = cart.CartCount
	rst["is_presale"] = cart.IsPresale
	rst["instant_rebate_money"] = cart.InstantRebateMoney
	rst["coupon_rebate_money"] = cart.CouponRebateMoney
	rst["total_rebate_money"] = cart.TotalRebateMoney
	rst["used_balance_money"] = cart.UsedBalanceMoney
	rst["can_used_balance_money"] = cart.CanUsedBalanceMoney
	rst["used_point_num"] = cart.UsedPointNum
	rst["used_point_money"] = cart.UsedPointMoney
	rst["can_used_point_num"] = cart.CanUsedPointNum
	rst["can_used_point_money"] = cart.CanUsedPointMoney
	rst["is_share_station"] = cart.IsShareStation
	rst["only_today_products"] = cart.OnlyTodayProducts
	rst["only_tomorrow_products"] = cart.OnlyTomorrowProducts
	rst["package_type"] = cart.PackageType
	rst["package_id"] = cart.PackageId
	rst["front_package_text"] = cart.FrontPackageText
	rst["front_package_type"] = cart.FrontPackageType
	rst["front_package_stock_color"] = cart.FrontPackageStockColor
	rst["front_package_bg_color"] = cart.FrontPackageBgColor
	return rst
}

func customizeReserve() map[string]int64 {
	switch reserve {
	case 2:
		return map[string]int64{"reserved_time_start": unix(14, 30), "reserved_time_end": unix(22, 30)}
	default:
		return map[string]int64{"reserved_time_start": unix(6, 30), "reserved_time_end": unix(14, 30)}
	}
}

func unix(hour, min int) int64 {
	now := time.Now()
	parse, _ := time.Parse("2006-01-02 15:04:05", time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.Local).String())
	return parse.Unix()
}

func reserveTime(products interface{}, aid string) map[string]int64 {

	if products == nil || aid == "" {
		return nil
	}

	marshal, _ := json.Marshal(products)
	client := http.Client{Timeout: time.Second * 5, Transport: tr}

	user := userInfo()
	user["addressId"] = []string{aid}
	user["products"] = []string{fmt.Sprintf("[%s]", string(marshal))}
	user["group_config_id"] = []string{""}
	user["isBridge"] = []string{"false"}
	const _url = "https://maicai.api.ddxq.mobi/order/getMultiReserveTime"
	req, _ := http.NewRequest(http.MethodPost, _url, strings.NewReader(user.Encode()))
	m := headers()
	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	req.Header = m

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	all, _ := ioutil.ReadAll(resp.Body)

	info := map[string]interface{}{}

	if err = json.Unmarshal(all, &info); err != nil {
		log.Println(err)
		return nil
	}

	if !httpStatus(info, "更新配送时间") {
		return nil
	}

	marshal, _ = json.Marshal(info["data"])
	var infos []Reserved
	if err = json.Unmarshal(marshal, &infos); err != nil {
		log.Println(err)
		return nil
	}

	for _, item := range infos[0].Time[0].Times {
		if item.DisableType == 0 {
			reserved := make(map[string]int64, 3)
			reserved["reserved_time_start"] = item.StartTimestamp
			reserved["reserved_time_end"] = item.EndTimestamp
			// log.Println("更新配送时间成功 start ", item.StartTimestamp, " end ", item.EndTimestamp)
			return reserved
		}
	}

	log.Println("配送时间已约满")

	return nil
}

func checkOrder(aid string, cart map[string]interface{}, reserve map[string]int64) map[string]interface{} {

	if aid == "" {
		return nil
	}

	infos := map[string]interface{}{}
	infos["products"] = cart["products"]
	infos["total_money"] = cart["total_money"]
	infos["total_origin_money"] = cart["total_money"]
	infos["goods_real_money"] = cart["goods_real_money"]
	infos["total_count"] = cart["total_count"]
	infos["cart_count"] = cart["cart_count"]
	infos["is_presale"] = cart["is_presale"]
	infos["instant_rebate_money"] = cart["instant_rebate_money"]
	infos["coupon_rebate_money"] = cart["coupon_rebate_money"]
	infos["total_rebate_money"] = cart["total_rebate_money"]
	infos["used_balance_money"] = cart["used_balance_money"]
	infos["can_used_balance_money"] = cart["can_used_balance_money"]
	infos["used_point_num"] = cart["used_point_num"]
	infos["used_point_money"] = cart["used_point_money"]
	infos["can_used_point_num"] = cart["can_used_point_num"]
	infos["can_used_point_money"] = cart["can_used_point_money"]
	infos["is_share_station"] = cart["is_share_station"]
	infos["only_today_products"] = cart["only_today_products"]
	infos["only_tomorrow_products"] = cart["only_tomorrow_products"]
	infos["package_type"] = cart["package_type"]
	infos["package_id"] = cart["package_id"]
	infos["front_package_text"] = cart["front_package_text"]
	infos["front_package_type"] = cart["front_package_type"]
	infos["front_package_stock_color"] = cart["front_package_stock_color"]
	infos["front_package_bg_color"] = cart["front_package_bg_color"]

	infos["reserved_time"] = map[string]interface{}{
		"reserved_time_start": reserve["reserved_time_start"],
		"reserved_time_end":   reserve["reserved_time_end"],
	}

	marshal, _ := json.Marshal(infos)

	request := userInfo()
	request["addressId"] = []string{aid}
	request["user_ticket_id"] = []string{"default"}
	request["freight_ticket_id"] = []string{"default"}
	request["is_use_point"] = []string{"0"}
	request["is_use_balance"] = []string{"0"}
	request["is_buy_vip"] = []string{"0"}
	request["coupons_id"] = []string{""}
	request["is_buy_coupons"] = []string{"0"}
	request["check_order_type"] = []string{"0"}
	request["is_support_merge_payment"] = []string{"1"}
	request["showData"] = []string{"true"}
	request["showMsg"] = []string{"false"}
	request["packages"] = []string{fmt.Sprintf("[%s]", string(marshal))}

	const _url = "https://maicai.api.ddxq.mobi/order/checkOrder"
	req, err := http.NewRequest(http.MethodPost, _url, strings.NewReader(request.Encode()))
	if err != nil {
		return nil
	}
	header := headers()
	req.Header = header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;UTF-8")

	client := http.Client{Timeout: 5 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		// log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	all, _ := ioutil.ReadAll(resp.Body)
	info := map[string]interface{}{}
	if err = json.Unmarshal(all, &info); err != nil {
		log.Println(err)
		return nil
	}

	if !httpStatus(info, "更新订单确认信息") {
		return nil
	}

	marshal, _ = json.Marshal(info["data"])
	order := Orders{}
	if err = json.Unmarshal(marshal, &order); err != nil {
		log.Println(err)
		return nil
	}

	orders := make(map[string]interface{}, 5)
	orders["freight_discount_money"] = order.Order.FreightDiscountMoney
	orders["freight_money"] = order.Order.FreightMoney
	orders["total_money"] = order.Order.TotalMoney
	orders["freight_real_money"] = order.Order.Freights[0].Freight.FreightRealMoney
	orders["user_ticket_id"] = order.Order.DefaultCoupon.ID
	log.Println("更新订单确认信息成功")
	return orders
}

func submitOrder(aid string, cart map[string]interface{}, order map[string]interface{}, reserve map[string]int64) bool {
	const _url = "https://maicai.api.ddxq.mobi/order/addNewOrder"

	request := userInfo()
	request["showMsg"] = []string{"false"}
	request["showData"] = []string{"true"}
	request["ab_config"] = []string{`{"key_onion":"C"}`}

	paymentOrder := map[string]interface{}{"reserved_time_start": reserve["reserved_time_start"],
		"reserved_time_end":      reserve["reserved_time_end"],
		"price":                  order["total_money"],
		"freight_discount_money": order["freight_discount_money"],
		"freight_money":          order["freight_money"],
		"order_freight":          order["freight_real_money"],
		"parent_order_sign":      cart["parent_order_sign"],
		"product_type":           1,
		"address_id":             aid,
		"form_id":                strconv.FormatInt(time.Now().UnixNano(), 10),
		"receipt_without_sku":    "",
		"pay_type":               6,
		"user_ticket_id":         order["user_ticket_id"],
		"vip_money":              "",
		"vip_buy_user_ticket_id": "",
		"coupons_money":          "",
		"coupons_id":             ""}

	packagesMap := []map[string]interface{}{
		{
			"products":                  cart["products"],
			"total_money":               cart["total_money"],
			"total_origin_money":        cart["total_money"],
			"goods_real_money":          cart["goods_real_money"],
			"total_count":               cart["total_count"],
			"cart_count":                cart["cart_count"],
			"is_presale":                cart["is_presale"],
			"instant_rebate_money":      cart["instant_rebate_money"],
			"coupon_rebate_money":       cart["coupon_rebate_money"],
			"total_rebate_money":        cart["total_rebate_money"],
			"used_balance_money":        cart["used_balance_money"],
			"can_used_balance_money":    cart["can_used_balance_money"],
			"used_point_num":            cart["used_point_num"],
			"used_point_money":          cart["used_point_money"],
			"can_used_point_num":        cart["can_used_point_num"],
			"can_used_point_money":      cart["can_used_point_money"],
			"is_share_station":          cart["is_share_station"],
			"only_today_products":       cart["only_today_products"],
			"only_tomorrow_products":    cart["only_tomorrow_products"],
			"package_type":              cart["package_type"],
			"package_id":                cart["package_id"],
			"front_package_text":        cart["front_package_text"],
			"front_package_type":        cart["front_package_type"],
			"front_package_stock_color": cart["front_package_stock_color"],
			"front_package_bg_color":    cart["front_package_bg_color"],
			"eta_trace_id":              "",
			"reserved_time_start":       reserve["reserved_time_start"],
			"reserved_time_end":         reserve["reserved_time_end"],
			"soon_arrival":              "",
			"first_selected_big_time":   1,
		},
	}

	packageOrder := map[string]interface{}{
		"payment_order": paymentOrder,
		"packages":      packagesMap,
	}

	marshal, _ := json.Marshal(packageOrder)
	request["package_order"] = []string{string(marshal)}

	req, err := http.NewRequest(http.MethodPost, _url, strings.NewReader(request.Encode()))
	if err != nil {
		log.Println(err)
		return false
	}
	header := headers()
	header["Content-Type"] = []string{"application/x-www-form-urlencoded;UTF-8"}
	req.Header = header

	client := http.Client{Timeout: time.Second * 5, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		// log.Println(err)
		return false
	}
	defer resp.Body.Close()
	all, _ := ioutil.ReadAll(resp.Body)
	info := map[string]interface{}{}
	if err = json.Unmarshal(all, &info); err != nil {
		log.Println(err)
		return false
	}

	if !httpStatus(info, "下单信息") {
		return false
	}

	rest := parseRest(info["data"])
	if len(rest) == 0 {
		return false
	}

	payUrl, ok := rest["pay_url"]
	if ok {
		if len(payUrl.(string)) != 0 {
			log.Println("成功下单 当前下单总金额：", cart["total_money"])
			return true
		}
		log.Println("下单失败")
	}
	return false
}

// httpStatus judge http status
func httpStatus(info map[string]interface{}, str string) bool {
	data, ok := info["success"].(bool)
	if !ok {
		return false
	}
	if !data {
		msg, ok := info["message"].(string)
		if ok {
			if "您的访问已过期" == msg {
				log.Println("用户信息失效")
			}
		} else {
			log.Println(str, "失败:", info["msg"])
		}
		return false
	}
	return true
}

// parseRest parse result data
func parseRest(body interface{}) map[string]interface{} {
	marshal, _ := json.Marshal(body)
	result := map[string]interface{}{}
	if err := json.Unmarshal(marshal, &result); err != nil {
		log.Println(err)
		return nil
	}
	return result
}
