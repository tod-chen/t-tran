package web

// IUser 用户行为接口
type IUser interface{
	// 查询车次及余票数
	queryTrans()
	// 提交订单
	submitOrder()
	// 查询订单
	queryOrder()
	// 取消订单
	cancelOrder()
	// 支付订单
	paymentOrder()
	// 退票
	refundOrder()
	// 改签
	changeOrder()
	// 取票
	printTicket()
}