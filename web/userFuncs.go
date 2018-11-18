package web

import (
	"net/http"
	"t-tran/modules"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// queryTimeout 用户发起请求的超时时间
	queryTimeout time.Duration = 5 * time.Second
)

func setUserRouter(g *gin.RouterGroup) {
	// 登录
	g.POST("/login", login)
	// 登出
	g.POST("/logout", logout)
	// 车次及余票查询
	g.GET("/residualTicket", queryResidualTicket)
}

func login(c *gin.Context) {

}

func logout(c *gin.Context) {

}

// queryTranResult 查询车次及余票结果
type queryTranResult struct {
	depStations map[string]string
	arrStations map[string]string
	tranInfos   []string
}

// 查询车次及余票数
func queryResidualTicket(c *gin.Context) {
	depStationName, arrStationName := c.Query("from"), c.Query("to")
	date, isStudent := c.Query("date"), c.DefaultQuery("isStudent", "0")
	c.JSON(http.StatusOK, gin.H{"trans": modules.QueryResidualTicketInfo(depStationName, arrStationName, date, isStudent == "1")})
}

// 查询时刻表
func queryTimetable(tranNum string, date time.Time) []modules.TimetableResult {
	return modules.QueryTimetable(tranNum, date)
}

// 查询票价
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
