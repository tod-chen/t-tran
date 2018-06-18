package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"t-tran/modules"

	"github.com/gin-gonic/gin"
)

func setAdminRoute(g *gin.Engine) {
	tranRouter(g)
	carRouter(g)
	scheduleRouter(g)
	stationRouter(g)
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

//////////////////////////////////////////////////
///            车次路由及方法               ///////
//////////////////////////////////////////////////
func tranRouter(g *gin.Engine) {
	g.GET("/trans", trans)
	g.GET("/trans/query", queryTrans)
	g.GET("/trans/detail", tranDetail)
	g.GET("/trans/getDetail", tranGetDetail)
	g.POST("/tran/save", saveTran)
}

func trans(c *gin.Context) {
	c.HTML(http.StatusOK, "trans.html", gin.H{})
}

func queryTrans(c *gin.Context) {
	tranNum, tranType := c.Query("tranNum"), c.Query("tranType")
	page, pageSize := getPaging(c)
	trans, count := modules.QueryTrans(tranNum, tranType, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"trans": trans, "count": count, "page": page, "ps": pageSize})
}

func tranDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "tranDetail.html", gin.H{})
}

func tranGetDetail(c *gin.Context) {
	tranID := c.DefaultQuery("tranId", "0")
	iTranID := strToInt(tranID, 0)
	tranInfo := modules.GetTranDetail(iTranID)
	c.JSON(http.StatusOK, gin.H{"tranInfo": tranInfo})
}

func saveTran(c *gin.Context) {
	var t modules.TranInfo
	c.BindJSON(&t)
	fmt.Println(t)
	c.JSON(http.StatusOK, gin.H{})
}

//////////////////////////////////////////////////
///            车厢路由及方法               ///////
//////////////////////////////////////////////////
func carRouter(g *gin.Engine) {
	g.GET("/cars", cars)
	g.GET("/cars/query", queryCars)
	g.GET("/cars/detail", carDetail)
	g.GET("/cars/getDetail", carGetDetail)
	g.POST("/car/save", saveCar)
}

func cars(c *gin.Context) {
	c.HTML(http.StatusOK, "cars.html", gin.H{})
}

func queryCars(c *gin.Context) {
	seatType, tranType := c.Query("seatType"), c.Query("tranType")
	page, pageSize := getPaging(c)
	cars, count := modules.QueryCars(seatType, tranType, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"cars": cars, "count": count, "page": page, "ps": pageSize})
}

func carDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "carDetail.html", gin.H{})
}

func carGetDetail(c *gin.Context) {
	carID := c.DefaultQuery("carId", "0")
	iCarID := strToInt(carID, 0)
	car := modules.GetCarDetail(iCarID)
	c.JSON(http.StatusOK, gin.H{"car": car})
}

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

//////////////////////////////////////////////////
///            排班路由及方法               ///////
//////////////////////////////////////////////////
func scheduleRouter(r *gin.Engine) {
	r.GET("/schedules", schedules)
	r.GET("/schedules/query", scheduleQuery)
	r.GET("/schedules/detail", scheduleDetail)
	r.GET("/schedules/getDetail", scheduleGetDetail)
	r.POST("/schedules/save", saveSchedule)
}
func schedules(c *gin.Context) {
	c.HTML(http.StatusOK, "schedules.html", gin.H{})
}

func scheduleQuery(c *gin.Context) {
	departureDate, tranNum := c.Query("depDate"), c.Query("tranNum")
	page, pageSize := getPaging(c)
	schedules, count := modules.QuerySchedule(departureDate, tranNum, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"schedules": schedules, "count": count, "page": page, "ps": pageSize})
}

func scheduleDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "scheduleDetail.html", gin.H{})
}

func scheduleGetDetail(c *gin.Context) {
	scheduleID := c.Query("scheduleID")
	iScheduleID := strToInt(scheduleID, 0)
	schedule := modules.GetScheduleDetail(iScheduleID)
	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func saveSchedule(c *gin.Context) {
	c.HTML(http.StatusOK, "schedules.html", gin.H{})
}

//////////////////////////////////////////////////
///            车站路由及方法               ///////
//////////////////////////////////////////////////
func stationRouter(r *gin.Engine) {
	r.GET("/stations", stations)
	r.GET("/stations/query", stationQuery)
	r.GET("/stations/detail", stationDetail)
	r.GET("/stations/getDetail", stationGetDetail)
	r.POST("/station/save", saveStation)
}

func stations(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}

func stationQuery(c *gin.Context) {
	stationName, cityName := c.Query("stationName"), c.Query("cityName")
	page, pageSize := getPaging(c)
	stations, count := modules.QueryStations(stationName, cityName, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"stations": stations, "count": count, "page": page, "ps": pageSize})
}

func stationDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "stationDetail.html", gin.H{})
}

func stationGetDetail(c *gin.Context) {
	stationID := c.DefaultQuery("stationID", "0")
	iStationID := strToInt(stationID, 0)
	station := modules.GetStationDetail(iStationID)
	c.JSON(http.StatusOK, gin.H{"station": station})
}

func saveStation(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}
