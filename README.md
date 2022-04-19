# dingdong 快到我的碗里来 😁

中文 | [English](./README_EN.md)

1. 这里使用burp suite 抓包叮咚小程序，我这使用mac，其它自行搜索
   1. [download](https://portswigger.net/burp/releases/professional-community-2022-2-4?requestededition=community)
   2. 打开burp suite 下载证书更名为ca.der 
   3. 导入burp suite 证书到本地
   4. 向burp suite 代理设置wifi或者其它软件
      ![set proxy](images/wifi.png)
2. 打开微信叮咚小程序和burp suite 
3. 打开burp suite proxy 按钮 查看 http history 历史 复制 token和headers
   1. 点击叮咚小程序购物车并在burp suite 查看接口 `/cart/index`
      copy url query param and headers info 复制 url的参数 和header info
      ![copy](images/cart_api.png)
   2. 在 go 代码函数 headers() 和 userInfo() 填写信息
4. 购物车准备好并且全选
5. start `go run .`
6. 下单成功自行支付，让我们祝疫情早日结束 🍻！！

# 实战

![dingdong_1](images/dingdong1.png)

![dingdong_2](images/dingdong2.png)

# feature
1. 新增运行程序，运行命令如下

   ./dingdong `-f config.yaml -aid addressid`

   `aid` 运行为空，默认去查找地址id
   
   `f` 向配置文件添加 URL query param和header info 需要的信息

2. 自定义预定时间抢单

   (1) 6:30-14:30

   (2) 14:30-22:00

   `reserve` 运行为空，默认是1