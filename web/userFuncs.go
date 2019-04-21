package web

import (
	"net/http"
	"strconv"
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
	// 查询车次及余票
	g.GET("/residualTicket", queryResidualTicket)
	// 查询时刻表
	g.GET("/queryTimetable", queryTimetable)
	// 查询票价
	g.GET("/queryPrice", queryPrice)
	// 提交订单
	g.POST("/submitOrder", submitOrder)
	// 确认改签
	g.POST("/changeOrder", changeOrder)
	// 查询订单
	g.GET("/queryOrder", queryOrder)
	// 取消订单
	g.POST("/cancelOrder", cancelOrder)
	// 退票
	g.POST("/refundOrder", refundOrder)
	// 出票
	g.POST("/printTicket", printTicket)

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
func queryTimetable(c *gin.Context) {
	tranNum, date := c.Query("tranNum"), c.Query("date")
	t, _ := time.ParseInLocation(modules.ConstYmdFormat, date, localLoc)
	c.JSON(http.StatusOK, gin.H{"timetable": modules.QueryTimetable(tranNum, t)})
}

// 查询票价
func queryPrice(c *gin.Context) {
	tranNum, date := c.Query("tranNum"), c.Query("date")
	t, _ := time.ParseInLocation(modules.ConstYmdFormat, date, localLoc)
	dep, arr := c.Query("depIdx"), c.Query("arrIdx")
	depI, _ := strconv.Atoi(dep)
	arrI, _ := strconv.Atoi(arr)
	depIdx, arrIdx := uint8(depI), uint8(arrI)
	c.JSON(http.StatusOK, gin.H{"price": modules.QuerySeatPrice(tranNum, t, depIdx, arrIdx)})
}

// 提交订单
func submitOrder(c *gin.Context) {
	var model modules.SubmitOrderModel
	if err := c.BindJSON(&model); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Post Data Err"})
		return
	}
	if err := modules.SubmitOrder(model); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 改签
func changeOrder(c *gin.Context) {
	var model modules.SubmitOrderModel
	if err := c.BindJSON(&model); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Post Data Err"})
		return
	}
	oldID := c.PostForm("oldTicketID")
	oldTicketID, err := strconv.ParseUint(oldID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "原车票信息无效"})
		return
	}
	if err = modules.ChangeOrder(model, oldTicketID); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 查询订单
func queryOrder(c *gin.Context) {

}

// 取消订单
func cancelOrder(c *gin.Context) {
	orderID := c.PostForm("orderID")
	oID, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "订单无效"})
		return
	}
	if err = modules.CancelOrder(oID); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 退票
func refundOrder(c *gin.Context) {
	orderID := c.PostForm("orderID")
	oID, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "订单无效"})
		return
	}
	if err = modules.RefundOrder(oID); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 取票
func printTicket(c *gin.Context) {
	ticketID := c.PostForm("ticketID")
	tID, err := strconv.ParseUint(ticketID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "订单无效"})
		return
	}
	modules.CheckIn(tID)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
