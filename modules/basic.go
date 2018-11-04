package modules

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// 时间格式
	constHmFormat     = "15:04"               // 时间格式 小时:分钟
	constHmsFormat    = "15:04:05"            // 时间格式 小时:分钟:秒
	constYmdFormat    = "2006-01-02"          // 时间格式 年-月-日
	constYMdHmFormat  = "2006-01-02 15:04"    // 时间格式 年-月-日 小时:分钟
	constYMdHmsFormat = "2006-01-02 15:04:05" // 时间格式 年-月-日 小时:分钟:秒
	constStrNullTime  = "----"

	constOneDayDuration = 24 * time.Hour // 一天的长度
	constTranCount      = 13000          // 车次数量
	constCityCount      = 620            // 有火车经过的城市数量
	// 各类座位的名称
	constSeatTypeSpecial             = "S"   // 商务座
	constSeatTypeFristClass          = "FC"  // 一等座
	constSeatTypeSecondClass         = "SC"  // 二等座
	constSeatTypeAdvancedSoftSleeper = "ASS" // 高级软卧
	constSeatTypeSoftSleeper         = "SS"  // 软卧
	constSeatTypeEMUSleeper          = "DS"  // 动车组卧铺
	constSeatTypeMoveSleeper         = "MS"  // 动卧(普快、直达、特快等车次，下铺床位改座位)
	constSeatTypeHardSleeper         = "HS"  // 硬卧
	constSeatTypeSoftSeat            = "SST" // 软座
	constSeatTypeHardSeat            = "HST" // 硬座
	constSeatTypeNoSeat              = "NST" // 无座
)

var (
	// 所有车厢，存于内存，便于组装
	carMap map[int](Car)
	// 所有车次信息，不参与订票，用于查询列车的时刻表和各路段各座次的价格
	tranInfos []TranInfo
	// 各城市与经过该城市的列车映射
	cityTranMap map[string]([]*TranInfo)
)

func initTranInfo() {
	initCarMap()
	initTranInfos()
	initCityTranMap()
}

func initCarMap() {
	fmt.Println("init car map begin")
	defer fmt.Println("init car map end")
	var cars []Car
	db.Find(&cars)
	fmt.Println("cars count:", len(cars))
	for i := 0; i < len(cars); i++ {
		db.Where("car_id = ?", cars[i].ID).Find(&cars[i].Seats)
		carMap[cars[i].ID] = cars[i]
	}
}
func initTranInfos() {
	fmt.Println("init Tran Info begin")
	defer fmt.Println("init Tran Info end")
	now, lastDay := time.Now(), time.Now().AddDate(0, 0, constDays)
	today, lastDate := now.Format(constYmdFormat), lastDay.Format(constYmdFormat)
	db.Where("enable_end_date >= ? and ? >= enable_start_date", today, lastDate).Order("tran_num, enable_start_date").Find(&tranInfos)
	for i := 0; i < len(tranInfos); i++ {
		tranInfos[i].getFullInfo()
	}
}
func initCityTranMap() {
	fmt.Println("init City Map begin")
	defer fmt.Println("init City Map end")
	cityTranMap = make(map[string]([]*TranInfo), constCityCount)
	for i := 0; i < len(tranInfos); i++ {
		for j := 0; j < len(tranInfos[i].Timetable); j++ {
			cityCode := tranInfos[i].Timetable[j].CityCode
			tranPtrs, exist := cityTranMap[cityCode]
			if exist {
				tranPtrs = append(tranPtrs, &tranInfos[i])
			} else {
				tranPtrs = []*TranInfo{&tranInfos[i]}
			}
			cityTranMap[cityCode] = tranPtrs
		}
	}
}
func getTranInfo(tranNum string, date time.Time) *TranInfo {
	idx := sort.Search(len(tranInfos), func(i int) bool {
		return tranInfos[i].TranNum == tranNum &&
			tranInfos[i].EnableStartDate.Before(date) &&
			tranInfos[i].EnableEndDate.After(date)
	})
	return &tranInfos[idx]
}
func getViaTrans(depS, arrS *Station) (result []*TranInfo){
	// 获取经过出发站所在城市的所有车次
	depTrans, exist := cityTranMap[depS.CityCode]
	if !exist {
		return
	}
	// 获取经过目的站所在城市的所有车次
	arrTrans, exist := cityTranMap[arrS.CityCode]
	if !exist{
		return
	}
	if len(arrTrans) < len(depTrans){
		arrTrans, depTrans = depTrans, arrTrans
	}
	// 用map取经过出发站城市车次和目的站城市车次的交集
	m := make(map[string](bool), len(depTrans))
	for _, t := range depTrans{
		m[t.TranNum + t.EnableStartDate.Format(constYmdFormat)] = false
	}
	for idx, t := range arrTrans{
		// 属于交集，则放置结果集中
		if _, ok := m[t.TranNum + t.EnableStartDate.Format(constYmdFormat)]; ok {
			result = append(result, arrTrans[idx])
		}
	}
	return
}

// TranInfo 列车信息结构体 ===============================start
type TranInfo struct {
	TranNum              string    `gorm:"index:main;type:varchar(10)" json:"tranNum"` // 车次号
	RouteDepDurationDays int       `json:"durationDays"`                               // 路段出发间隔天数：最后一个路段的发车时间与起点站发车时间的间隔天数
	ScheduleDays         int       `gorm:"default:1" json:"scheduleDays"`              // 间隔多少天发一趟车，绝大多数是1天
	IsSaleTicket         bool      `gorm:"default:1"`                                  // 是否售票
	SaleTicketTime       time.Time // 售票时间，不需要日期部分，只取时间部分
	NonSaleRemark        string    `gorm:"type:varchar(100)"` // 不售票说明
	// 生效开始日期 默认零值
	EnableStartDate time.Time `gorm:"index:query;type:datetime;default:'0000-00-00 00:00:00'" json:"enableStartDate"`
	// 生效截止日期 默认最大值
	EnableEndDate time.Time              `gorm:"index:query;type:datetime;default:'9999-12-31 23:59:59'" json:"enableEndDate"`
	CarIds        string                 `gorm:"type:varchar(100)" json:"carIds"` // 车厢ID及其数量，格式如：32:1;12:2; ...
	Timetable     []Route                `gorm:"-" json:"timetable"`              // 时刻表
	SeatPriceMap  map[string]([]float32) `gorm:"-" json:"seatPriceMap"`           // 各类席位在各路段的价格
	DBModel
}

// 是否为城际车次，城际车次在同一个城市内可能会有多个站，情况相对特殊
func (t *TranInfo) isIntercity() bool {
	return strings.Index(t.TranNum, "C") == 0
}
func (t *TranInfo) getFullInfo() {
	// 获取时刻表信息
	var routes []Route
	db.Where("tran_id = ?", t.ID).Order("station_index").Find(&routes)
	t.Timetable = routes

	// 获取各席别在各路段的价格，大多数车次只有三类席别（无座不考虑）
	t.SeatPriceMap = make(map[string]([]float32), 3)
	var routePrices []RoutePrice
	db.Where("tran_id = ?", t.ID).Order("seat_type, route_index").Find(&routePrices)
	for start, end := 0, 1; end < len(routePrices); end++ {
		if end == len(routePrices)-1 || routePrices[start].SeatType != routePrices[end].SeatType {
			arr := routePrices[start:end]
			prices := make([]float32, len(arr))
			for idx, rp := range arr {
				prices[idx] = rp.Price
			}
			t.SeatPriceMap[routePrices[start].SeatType] = prices
			start = end
		}
	}
}
func (t *TranInfo) getScheduleCars() (sCars *[]ScheduleCar) {
	// 获取车厢信息
	carSettings, carCount := strings.Split(t.CarIds, ";"), 0
	carIds, carIDCountMap := make([]int, len(carSettings)), make(map[int]int, len(carSettings))
	for i := 0; i < len(carSettings); i++ {
		setting := strings.Split(carSettings[i], ":")
		if len(setting) == 2 {
			id, _ := strconv.Atoi(setting[0])
			carIds[i] = id
			count, _ := strconv.Atoi(setting[1])
			carIDCountMap[id] = count
			carCount += count
		}
	}
	result, carIdx := make([]ScheduleCar, carCount), uint8(0)
	for id, count := range carIDCountMap {
		c := carMap[id]
		for i:=0; i< count;i++{
			result[carIdx] = ScheduleCar{
				SeatType:    c.SeatType,
				CarNum:      uint8(len(result)) + 1,
				NoSeatCount: c.NoSeatCount,
				Seats:       make([]ScheduleSeat, len(c.Seats)),
				EachRouteTravelerCount: make([]uint8, len(t.Timetable)-1),
			}
			for si :=0; si<len(c.Seats);si++{
				result[carIdx].Seats[si].SeatNum = c.Seats[si].SeatNum
				result[carIdx].Seats[si].IsStudent = c.Seats[si].IsStudent
			}
			carIdx++
		}
	}
	return &result
}

// Save 保存到数据库
func (t *TranInfo) Save() (bool, string) {
	t.initTimetable()
	t.EnableEndDate = t.EnableEndDate.Add(24*time.Hour - time.Second)
	if t.ID == 0 {
		db.Create(t)
	} else {
		db.Save(t)
		db.Delete(Route{}, "tran_id = ?", t.ID)
		db.Delete(RoutePrice{}, "tran_id = ?", t.ID)
	}
	for i, r := range t.Timetable {
		r.TranID = t.ID
		r.TranNum = t.TranNum
		r.StationIndex = uint8(i + 1)
		db.Create(&r)
	}
	for k, v := range t.SeatPriceMap {
		for i, p := range v {
			rp := &RoutePrice{TranID: t.ID, SeatType: k, RouteIndex: uint8(i), Price: p}
			db.Create(rp)
		}
	}
	return true, ""
}

// 重置列车所经站点的时间，默认从0001-01-01开始，终点站的到达时间0001-01-0N, 其中‘N’表示该列车运行的跨天数
func (t *TranInfo) initTimetable() {
	// 所跨天数
	day, routeCount := 0, len(t.Timetable)
	// Note: 这里默认有一个规则，所有列车在任一路段，运行时间不会超过24小时；当有例外时，下面的代码需要调整逻辑
	// 起点站无需重置出发时间和到达时间，终点站无需重置出发时间
	for i := 1; i < routeCount; i++ {
		t.Timetable[i].ArrTime = t.Timetable[i].ArrTime.AddDate(0, 0, day)
		if t.Timetable[i].ArrTime.Before(t.Timetable[i-1].DepTime) {
			day++
			t.Timetable[i].ArrTime = t.Timetable[i].ArrTime.AddDate(0, 0, 1)
		}
		if i == routeCount-1 {
			break
		}
		t.Timetable[i].DepTime = t.Timetable[i].DepTime.AddDate(0, 0, day)
		if t.Timetable[i].DepTime.Before(t.Timetable[i].ArrTime) {
			day++
			t.Timetable[i].DepTime = t.Timetable[i].DepTime.AddDate(0, 0, 1)
		}
	}
	// 时刻表中的第一站出发日期，默认都为0001-01-01，所以最后一个路段的发车日期 - 1，就是路段出发间隔天数
	t.RouteDepDurationDays = t.Timetable[len(t.Timetable)-1].DepTime.YearDay() - 1
}

// 根据起止站获取各类座位的票价
func (t *TranInfo) getSeatPrice(depIdx, arrIdx uint8) (result map[string]float32) {
	for seatType, eachRoutePrice := range t.SeatPriceMap {
		var price float32
		for i := depIdx; i < arrIdx; i++ {
			price += eachRoutePrice[i]
		}
		result[seatType] = price
	}
	return
}

// IsMatchQuery 判断当前车次在站点及日期上是否匹配
func (t *TranInfo) IsMatchQuery(depS, arrS *Station, queryDate time.Time) (depIdx, arrIdx uint8, depDate string, ok bool) {
	depI := -1
	// 非城际车次
	if !t.isIntercity() {
		// Note: 当某车次的路线经过某城市的两个站，该怎么匹配？ 当前算法是匹配第一个，与12306逻辑一致 12306这里算是一个bug
		for i := 0; i < len(t.Timetable); i++ {
			if depI == -1 && t.Timetable[i].CityCode == depS.CityCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.Timetable[i].CityCode == arrS.CityCode {
				arrIdx = uint8(i)
				ok = true
				break
			}
		}
	} else { // 城际车次
		for i := 0; i < len(t.Timetable); i++ {
			if depI == -1 && t.Timetable[i].StationCode == depS.StationCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.Timetable[i].StationCode == arrS.StationCode {
				arrIdx = uint8(i)
				ok = true
				break
			}
		}
	}
	if !ok {
		return
	}
	// 计算当前车次信息的出发站发车日期
	date := queryDate.AddDate(0, 0, 1-t.Timetable[depIdx].DepTime.Day())
	// 该发车日期不在有效时间段内，则返回false
	if date.Before(t.EnableStartDate) || date.After(t.EnableEndDate) {
		ok = false
		return
	}
	ok = true
	// 不是每天发车的车次，要判断发车日期是否有效
	if t.ScheduleDays > 1 {
		if date.Sub(t.EnableStartDate).Hours()/float64(24*t.ScheduleDays) != 0 {
			ok = false
		}
	}
	depDate = date.Format(constYmdFormat)
	return
}

// Route 时刻表信息
type Route struct {
	TranID          uint64  // 车次ID
	TranNum         string  `gorm:"index:main;type:varchar(10)"` // 车次号
	StationIndex    uint8   // 车站索引
	StationName     string  `gorm:"type:nvarchar(20)" json:"stationName"`    // 车站名
	StationCode     string  `gorm:"type:varchar(10)"`                        // 车站编码
	CityCode        string  `gorm:"type:varchar(20)"`                        // 城市编码
	CheckTicketGate string  `gorm:"type:varchar(10)" json:"checkTicketGate"` // 检票口
	Platform        uint8   `json:"platform"`                                // 乘车站台
	MileageNext     float32 `json:"mileageNext"`                             // 距下一站的里程
	// 出发时间
	DepTime time.Time `gorm:"type:datetime" json:"depTime"`
	// 到达时间
	ArrTime time.Time `gorm:"type:datetime" json:"arrTime"`
	DBModel
}

func (r *Route) getStrDepTime() string {
	if r.DepTime.Year() == 1 {
		return constStrNullTime
	}
	return r.DepTime.Format(constHmFormat)
}

func (r *Route) getStrArrTime() string {
	if r.ArrTime.Year() == 1 {
		return constStrNullTime
	}
	return r.ArrTime.Format(constHmFormat)
}

func (r *Route) getStrStayTime() string {
	if r.DepTime.Year() == 1 || r.ArrTime.Year() == 1 {
		return constStrNullTime
	}
	return strconv.FormatFloat(r.DepTime.Sub(r.ArrTime).Minutes(), 'e', 0, 64)
}

// RoutePrice 各路段价格
type RoutePrice struct {
	TranID     uint64  `gorm:"index:main"`      // 车次ID
	SeatType   string  `gorm:"type:varchar(5)"` // 座次类型
	RouteIndex uint8   // 路段索引
	Price      float32 // 价格
	DBModel
}

// Car 车厢信息结构体
type Car struct {
	ID          int
	TranType    string `gorm:"type:varchar(20)" json:"tranType"` // 车次类型 高铁、动车、直达等
	SeatType    string `gorm:"type:varchar(5)" json:"seatType"`  // 车厢的座位类型
	SeatCount   uint8  // 车厢内座位数(或床位数)
	NoSeatCount uint8  `json:"noSeatCount"`                     // 车厢内站票数
	Remark      string `gorm:"type:nvarchar(50)" json:"remark"` // 说明
	Seats       []Seat `json:"seats"`                           // 车厢的所有座位
}

// Save 保存车厢信息到数据库
func (c *Car) Save() (bool, string) {
	c.SeatCount = uint8(len(c.Seats))
	if c.ID == 0 {
		db.Create(c)
	} else {
		db.Save(c)
		db.Delete(Seat{}, "car_id = ?", c.ID)
	}
	for i := 0; i < len(c.Seats); i++ {
		c.Seats[i].CarID = c.ID
		db.Create(&c.Seats[i])
	}
	return true, ""
}

// Seat 座位信息结构体
type Seat struct {
	CarID     int    `gorm:"index:main"`                     // 车厢ID
	SeatNum   string `gorm:"type:varchar(5)" json:"seatNum"` // 座位号
	IsStudent bool   `json:"isStudent"`                      // 是否学生票
}
