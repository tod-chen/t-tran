package modules

const (
	// PayTypeAliPay 支付类型-支付宝
	PayTypeAliPay = 1
	// PayTypeWeChatPay 支付类型-微信
	PayTypeWeChatPay = 2
)

// Payment 收款
func Payment(orderID, userID uint64, payType uint8, payAccount string, price float32, timestamp string) error {
	return nil
}

// Refund 退款
func Refund(orderID, userID uint64, payType uint8, payAccount string, price float32, timestamp string) error {
	return nil
}
