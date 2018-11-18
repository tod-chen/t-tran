package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"t-tran/modules"
	"time"

	"github.com/gin-gonic/gin"
)

func setAdminRouter(g *gin.RouterGroup) {
	// 车次路由
	g.GET("/trans", trans)
	g.GET("/trans/query", queryTrans)
	g.GET("/trans/detail", tranDetail)
	g.GET("/trans/getDetail", getTranDetail)
	g.POST("/tran/save", saveTran)

	// 车厢路由
	g.GET("/cars", cars)
	g.GET("/cars/query", queryCars)
	g.GET("/cars/detail", carDetail)
	g.GET("/cars/getDetail", getCarDetail)
	g.POST("/car/save", saveCar)

	// 车站路由
	g.GET("/stations", stations)
	g.GET("/stations/query", stationQuery)
	g.GET("/stations/detail", stationDetail)
	g.GET("/stations/getDetail", getStationDetail)
	g.POST("/station/save", saveStation)

	// 排班路由
	g.GET("/schedules", schedules)
	g.GET("/schedules/query", scheduleQuery)
	g.GET("/schedules/detail", scheduleDetail)
	g.GET("/schedules/getDetail", getScheduleDetail)
	g.POST("/schedules/save", saveSchedule)
}

func getPaging(c *gin.Context) (page, pageSize int) {
	pageSize = 15
	pageIdx := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageIdx)
	if err != nil || page < 1 {
		page = 1
	}
	return
}

func strToInt(str string, defVal int) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return defVal
	}
	return val
}

// trans 返回车次页面
func trans(c *gin.Context) {
	c.HTML(http.StatusOK, "trans.html", gin.H{})
}

// queryTrans 查询车次配置 & 翻页
func queryTrans(c *gin.Context) {
	tranNum, tranType := c.Query("tranNum"), c.Query("tranType")
	page, pageSize := getPaging(c)
	trans, count := modules.QueryTrans(tranNum, tranType, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"trans": trans, "count": count, "page": page, "ps": pageSize})
}

// tranDetail 返回车次配置详情页
func tranDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "tranDetail.html", gin.H{})
}

// getTranDetail 获取车次配置详情信息
func getTranDetail(c *gin.Context) {
	tranID := c.DefaultQuery("tranId", "0")
	iTranID := strToInt(tranID, 0)
	tranInfo := modules.GetTranDetail(iTranID)
	c.JSON(http.StatusOK, gin.H{"tranInfo": tranInfo})
}

// saveTran 保存车次配置信息
func saveTran(c *gin.Context) {
	var t modules.TranInfo
	if err := c.BindJSON(&t); err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(t.Timetable); i++ {
		t.Timetable[i].ArrTime.Add(8*time.Hour).AddDate(-1970, 0, 0)
		t.Timetable[i].DepTime.Add(8*time.Hour).AddDate(-1970, 0, 0)
	}
	success, msg := t.Save()
	c.JSON(http.StatusOK, gin.H{"success": success, "msg": msg})
}

// cars 返回车厢页面
func cars(c *gin.Context) {
	c.HTML(http.StatusOK, "cars.html", gin.H{})
}

// queryCars 查询车厢配置 & 翻页
func queryCars(c *gin.Context) {
	seatType, tranType := c.Query("seatType"), c.Query("tranType")
	page, pageSize := getPaging(c)
	cars, count := modules.QueryCars(seatType, tranType, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"cars": cars, "count": count, "page": page, "ps": pageSize})
}

// carDetail 返回车厢详情页
func carDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "carDetail.html", gin.H{})
}

// getCarDetail 获取车厢配置详细信息
func getCarDetail(c *gin.Context) {
	carID := c.DefaultQuery("carId", "0")
	iCarID := strToInt(carID, 0)
	car := modules.GetCarDetail(iCarID)
	c.JSON(http.StatusOK, gin.H{"car": car})
}

// saveCar 保存车厢配置信息
func saveCar(c *gin.Context) {
	var car modules.Car
	if err := c.BindJSON(&car); err != nil {
		log.Panicln(err)
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Post Data Err"})
		return
	}
	success, msg := car.Save()
	c.JSON(http.StatusOK, gin.H{"success": success, "msg": msg})
}

// stations 返回车站页
func stations(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}

// stationQuery 查询车站 & 翻页
func stationQuery(c *gin.Context) {
	stationName, cityName := c.Query("stationName"), c.Query("cityName")
	page, pageSize := getPaging(c)
	stations, count := modules.QueryStations(stationName, cityName, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"stations": stations, "count": count, "page": page, "ps": pageSize})
}

// stationDetail 返回车站详情页
func stationDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "stationDetail.html", gin.H{})
}

// getStationDetail 获取车站详细信息 
func getStationDetail(c *gin.Context) {
	stationID := c.DefaultQuery("stationID", "0")
	iStationID := strToInt(stationID, 0)
	station := modules.GetStationDetail(iStationID)
	c.JSON(http.StatusOK, gin.H{"station": station})
}

// saveStation 保存车站信息
func saveStation(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}

// schedules 返回排班页
func schedules(c *gin.Context) {
	c.HTML(http.StatusOK, "schedules.html", gin.H{})
}

// scheduleQuery 查询排班 & 翻页
func scheduleQuery(c *gin.Context) {
	departureDate, tranNum := c.Query("depDate"), c.Query("tranNum")
	page, pageSize := getPaging(c)
	schedules, count := modules.QuerySchedule(departureDate, tranNum, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"schedules": schedules, "count": count, "page": page, "ps": pageSize})
}

// scheduleDetail 返回排班详情页
func scheduleDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "scheduleDetail.html", gin.H{})
}

// getScheduleDetail 获取排班详细信息
func getScheduleDetail(c *gin.Context) {
	scheduleID := c.Query("scheduleID")
	iScheduleID := strToInt(scheduleID, 0)
	schedule := modules.GetScheduleDetail(iScheduleID)
	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

// saveSchedule 保存排班信息
func saveSchedule(c *gin.Context) {
	var schedule modules.ScheduleTran
	if err := c.BindJSON(&schedule); err != nil {
		log.Panicln(err)
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Post Data Err"})
		return
	}
	success, msg := schedule.Save()
	c.JSON(http.StatusOK, gin.H{"success": success, "msg": msg})
}