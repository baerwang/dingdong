package main

type Cart struct {
	CartCount              int           `json:"cart_count"`
	IsPresale              int           `json:"is_presale"`
	OnlyTomorrowProducts   []interface{} `json:"only_tomorrow_products"`
	FrontPackageType       int           `json:"front_package_type"`
	TotalCount             int           `json:"total_count"`
	GoodsRealMoney         string        `json:"goods_real_money"`
	CanUsedPointNum        int           `json:"can_used_point_num"`
	PackageId              int           `json:"package_id"`
	FrontPackageBgColor    string        `json:"front_package_bg_color"`
	PackageType            int           `json:"package_type"`
	CanUsedBalanceMoney    string        `json:"can_used_balance_money"`
	CanUsedPointMoney      string        `json:"can_used_point_money"`
	IsShareStation         int           `json:"is_share_station"`
	Products               []Product     `json:"products"`
	TotalOriginMoney       string        `json:"total_origin_money"`
	UsedBalanceMoney       string        `json:"used_balance_money"`
	OnlyTodayProducts      []interface{} `json:"only_today_products"`
	InstantRebateMoney     string        `json:"instant_rebate_money"`
	CouponRebateMoney      string        `json:"coupon_rebate_money"`
	FrontPackageStockColor string        `json:"front_package_stock_color"`
	FrontPackageText       string        `json:"front_package_text"`
	UsedPointNum           int           `json:"used_point_num"`
	TotalMoney             string        `json:"total_money"`
	TotalRebateMoney       string        `json:"total_rebate_money"`
	UsedPointMoney         string        `json:"used_point_money"`
}

type Orders struct {
	Order struct {
		DefaultCoupon struct {
			ID string `json:"_id"`
		} `json:"default_coupon"`
		DefaultFreightCoupon struct {
		} `json:"default_freight_coupon"`
		DisplayTotalMoney    string      `json:"display_total_money"`
		FreightDiscountMoney interface{} `json:"freight_discount_money"`
		FreightMoney         string      `json:"freight_money"`
		Freights             []struct {
			Freight struct {
				DiscountFreightMoney string `json:"discount_freight_money"`
				FreightMoney         string `json:"freight_money"`
				FreightRealMoney     string `json:"freight_real_money"`
				Remark               string `json:"remark"`
				Type                 int    `json:"type"`
			} `json:"freight"`
			PackageId int `json:"package_id"`
		} `json:"freights"`
		TotalMoney string `json:"total_money"`
	} `json:"order"`
}

type Product struct {
	ActivityId                string `json:"activity_id"`
	BuyLimit                  int    `json:"buy_limit"`
	CartId                    string `json:"cart_id"`
	CategoryPath              string `json:"category_path"`
	ConditionsNum             string `json:"conditions_num"`
	Count                     int    `json:"count"`
	DeliveryDateTag           string `json:"delivery_date_tag"`
	Description               string `json:"description"`
	Features                  string `json:"features"`
	Id                        string `json:"id"`
	InstantRebateMoney        string `json:"instant_rebate_money"`
	IsBooking                 int    `json:"is_booking"`
	IsBulk                    int    `json:"is_bulk"`
	IsGift                    int    `json:"is_gift"`
	IsInvoice                 int    `json:"is_invoice"`
	IsPresale                 int    `json:"is_presale"`
	IsSharedStationProduct    int    `json:"is_shared_station_product"`
	ManageCategoryPath        string `json:"manage_category_path"`
	NetWeight                 string `json:"net_weight"`
	NetWeightUnit             string `json:"net_weight_unit"`
	NoSupplementaryPrice      string `json:"no_supplementary_price"`
	NoSupplementaryTotalPrice string `json:"no_supplementary_total_price"`
	OrderSort                 int    `json:"order_sort"`
	OriginPrice               string `json:"origin_price"`
	ParentBatchType           int    `json:"parent_batch_type"`
	ParentId                  string `json:"parent_id"`
	Price                     string `json:"price"`
	PriceType                 int    `json:"price_type"`
	ProductName               string `json:"product_name"`
	ProductType               int    `json:"product_type"`
	PromotionNum              int    `json:"promotion_num"`
	SaleBatches               struct {
		BatchType int `json:"batch_type"`
	} `json:"sale_batches"`
	SizePrice         string        `json:"size_price"`
	Sizes             []interface{} `json:"sizes"`
	SkuActivityId     string        `json:"sku_activity_id"`
	SmallImage        string        `json:"small_image"`
	StorageValueId    int           `json:"storage_value_id"`
	SubList           []interface{} `json:"sub_list"`
	SupplementaryList []interface{} `json:"supplementary_list"`
	TemperatureLayer  string        `json:"temperature_layer"`
	TotalOriginPrice  string        `json:"total_origin_price"`
	TotalPrice        string        `json:"total_price"`
	Type              int           `json:"type"`
	ViewTotalWeight   string        `json:"view_total_weight"`
	TotalMoney        string        `json:"total_money"`
	TotalOriginMoney  string        `json:"total_origin_money"`
}

type Reserved struct {
	AreaLevel           int         `json:"area_level"`
	BusySoonArrivalText string      `json:"busy_soon_arrival_text"`
	DefaultSelect       bool        `json:"default_select"`
	EtaTraceId          string      `json:"eta_trace_id"`
	IntervalMinute      int         `json:"interval_minute"`
	IsNewRules          bool        `json:"is_new_rules"`
	StationDelayText    interface{} `json:"station_delay_text"`
	StationId           string      `json:"station_id"`
	Time                []struct {
		DateStr          string      `json:"date_str"`
		DateStrTimestamp int         `json:"date_str_timestamp"`
		Day              string      `json:"day"`
		InvalidPrompt    interface{} `json:"invalid_prompt"`
		IsInvalid        bool        `json:"is_invalid"`
		TimeFullTextTip  interface{} `json:"time_full_text_tip"`
		Times            []struct {
			ArrivalTime    bool   `json:"arrival_time"`
			ArrivalTimeMsg string `json:"arrival_time_msg"`
			DisableMsg     string `json:"disableMsg"`
			DisableType    int    `json:"disableType"`
			EndTime        string `json:"end_time"`
			EndTimestamp   int64  `json:"end_timestamp"`
			FullFlag       bool   `json:"fullFlag"`
			SelectMsg      string `json:"select_msg"`
			StartTime      string `json:"start_time"`
			StartTimestamp int64  `json:"start_timestamp"`
			TextMsg        string `json:"textMsg"`
			Type           int    `json:"type"`
		} `json:"times"`
	} `json:"time"`
}
