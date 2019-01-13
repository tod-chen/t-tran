package modules

import "time"

type notifyCustomerInfo struct {
	phoneNum        string    // 用户电话（短信提醒）
	emailAddr       string    // 用户邮箱（邮件提醒）
	orderNum        string    // 订单号
	userName        string    // 用户姓名
	passengerName   string    // 乘车人姓名
	tranNum         string    // 车次
	carNum          string    // 车厢号
	seatNum         string    // 座位号
	seatType        string    // 席别
	checkTicketGate string    // 检票口
	depTime         time.Time // 乘车日期
	depStation      string    // 出发站
	arrStation      string    // 到达站
}

type notifyAdminInfo struct {
	date       string // 日期
	tranNum    string // 车次
	carNum     uint8  // 车厢号
	depStation string // 出发站
	arrStation string // 到达站
	notifyType string // 通知类型
	message    string // 消息说明
}

func (n *notifyAdminInfo) notifyAdmin() {

}
