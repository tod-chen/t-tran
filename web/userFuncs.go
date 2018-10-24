package web

import (
	"t-tran/modules"
	"time"
)

const (
	// queryTimeout 用户发起请求的超时时间
	queryTimeout time.Duration = 5 * time.Second
)

// queryTranResult 查询车次及余票结果
type queryTranResult struct {
	depStations map[string]string
	arrStations map[string]string
	tranInfos   []string
}

// 查询车次及余票数
func queryResidualTicketInfo(depStationName, arrStationName string, depDate time.Time, isStudent bool) []string {
	return modules.QueryResidualTicketInfo(depStationName, arrStationName, depDate, isStudent)
}

// // 查询时刻表
func queryTimetable(tranNum string, date time.Time) []modules.TimetableResult {
	return modules.QueryTimetable(tranNum, date)
}

// // 查询票价
func queryPrice(tranNum string, date time.Time, depIdx, arrIdx uint8) map[string]float32 {
	return modules.QuerySeatPrice(tranNum, date, depIdx, arrIdx)
}

// 提交订单
func submitOrder() {

}

// 查询订单
func queryOrder() {

}

// 取消订单
func cancelOrder() {

}

// 支付订单
func paymentOrder() {

}

// 退票
func refundOrder() {

}

// 改签
func changeOrder() {

}

// 取票
func printTicket() {

}
