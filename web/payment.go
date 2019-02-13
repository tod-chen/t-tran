package web

import(
	"strconv"
	"t-tran/modules"

	"github.com/gin-gonic/gin"
)

func setPaymentRouter(g *gin.RouterGroup){
	// 支付宝支付回调
	g.POST("/alipay", alipayCallback)
	// 微信支付回调
	g.POST("/wechatpay", wechatpayCallback)
}

func alipayCallback(c *gin.Context){
	// 订单号
	orderID := c.PostForm("out_trade_no")
	// 支付宝流水号
	//_ := c.PostFrom("trade_no")
	// 支付金额
	totalAmount := c.PostForm("total_amount")
	// 支付方用户号
	sellerID := c.PostForm("seller_id")
	// 时间
	timestamp := c.PostForm("timestamp")
	// 结果码
	/*
	9000	订单支付成功
	8000	正在处理中，支付结果未知（有可能已经支付成功），请查询商户订单列表中订单的支付状态
	4000	订单支付失败
	5000	重复请求
	6001	用户中途取消
	6002	网络连接出错
	6004	支付结果未知（有可能已经支付成功），请查询商户订单列表中订单的支付状态
	*/
	code := c.PostForm("code")
	if code == "9000" {
		// 订单号
		oid, _ := strconv.ParseUint(orderID, 10, 64)
		order := modules.GetOrderInfo(oid)
		// 金额
		amount64, _ := strconv.ParseFloat(totalAmount, 64)
		amount := float32(amount64)
		modules.Payment(oid, order.UserID, modules.PayTypeAliPay, sellerID, amount, timestamp)
	}
}

func wechatpayCallback(c *gin.Context){

}