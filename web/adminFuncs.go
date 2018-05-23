package web

import (
	"net/http"
	"strconv"
	"t-tran/modules"

	"github.com/gin-gonic/gin"
)

func setAdminRoute(r *gin.Engine) {
	tranRouter(r)
	carRouter(r)
	scheduleRouter(r)
	stationRouter(r)
}

//////////////////////////////////////////////////
///            车次路由及方法               ///////
//////////////////////////////////////////////////
func tranRouter(r *gin.Engine) {
	r.GET("/trans", trans)
	r.GET("/trans/query", queryTrans)
	r.GET("/trans/detail", tranDetail)
	r.POST("/tran/save", saveTran)
}

func trans(c *gin.Context) {
	c.HTML(http.StatusOK, "trans.html", gin.H{})
}

func queryTrans(c *gin.Context) {
	tranNum, tranType := c.Param("tranNum"), c.Param("tranType")
	pageIdx, pageSize := c.DefaultQuery("page", "1"), 15
	page, err := strconv.Atoi(pageIdx)
	if err != nil {
		page = 1
	}
	trans := modules.QueryTrans(tranNum, tranType, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"trans": trans})
}

func tranDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "tranDetail.html", gin.H{})
}

func saveTran(c *gin.Context) {
	c.HTML(http.StatusOK, "trans.html", gin.H{})
}

//////////////////////////////////////////////////
///            车厢路由及方法               ///////
//////////////////////////////////////////////////
func carRouter(r *gin.Engine) {
	r.GET("/cars", cars)
	r.GET("/cars/detail")
	r.POST("/car/save", saveCar)
}

func cars(c *gin.Context) {
	c.HTML(http.StatusOK, "cars.html", gin.H{})
}

func carDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "carDetail.html", gin.H{})
}

func saveCar(c *gin.Context) {
	c.HTML(http.StatusOK, "cars.html", gin.H{})
}

//////////////////////////////////////////////////
///            排班路由及方法               ///////
//////////////////////////////////////////////////
func scheduleRouter(r *gin.Engine) {
	r.GET("/schedules", schedules)
	r.GET("/schedules/detail", scheduleDetail)
	r.POST("/schedules/save", saveSchedule)
}
func schedules(c *gin.Context) {
	c.HTML(http.StatusOK, "schedules.html", gin.H{})
}

func scheduleDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "scheduleDetail.html", gin.H{})
}

func saveSchedule(c *gin.Context) {
	c.HTML(http.StatusOK, "schedules.html", gin.H{})
}

//////////////////////////////////////////////////
///            车站路由及方法               ///////
//////////////////////////////////////////////////
func stationRouter(r *gin.Engine) {
	r.GET("/stations", stations)
	r.GET("/stations/detail", stationDetail)
	r.POST("/station/save", saveStation)
}

func stations(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}

func stationDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "stationDetail.html", gin.H{})
}

func saveStation(c *gin.Context) {
	c.HTML(http.StatusOK, "stations.html", gin.H{})
}
