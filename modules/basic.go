package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// 时间格式
	constHmFormat    = "15:04"            // 时间格式 小时:分钟
	constYmdFormat   = "2006-01-02"       // 时间格式 年-月-日
	constYMdHmFormat = "2006-01-02 15:04" // 时间格式 年-月-日 小时:分钟
	constStrNullTime = "----"

	constOneDayDuration          = 24 * time.Hour // 一天的长度
	constTranCount               = 13000          // 车次数量
	constDays                    = 30             // 可提前订票天数
	constCityCount               = 620            // 有火车经过的城市数量
	constQueryTranDelay          = 20             // 查询时距离当前时间多长时间内发车的车次不予显示 单位：分钟
	constUnpayOrderAvaliableTime = 45             // 未完成订单有效时间 单位：分钟
	constMaxAvaliableSeatCount   = 100            // 查询余票数量的最大值 超过此值时，显示“有”
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

func initTranInfo() {
	initTranList()
	initCityTranMap()
}

func initTranList() {
	fmt.Println("init tranList beginning")
	defer fmt.Println("init tranList end")

	today := time.Now().Format(constYmdFormat)
	fmt.Println("today format:", today)
	db.Where("enable_end_date >= ?", today).Find(&tranList)
	fmt.Println("tranList count:", len(tranList))
	for i := 0; i < len(tranList); i++ {
		tranList[i].getFullInfo()
	}
}

func initCityTranMap() {
	fmt.Println("init cityTranMap beginning")
	defer fmt.Println("init cityTranMap end")
	cityTranMap = make(map[string]([]*TranInfo), constCityCount)
	for i := 0; i < len(tranList); i++ {
		for j := 0; j < len(tranList[i].RouteTimetable); j++ {
			cityCode := tranList[i].RouteTimetable[j].CityCode
			tranPtrs, exist := cityTranMap[cityCode]
			if exist {
				tranPtrs = append(tranPtrs, &tranList[i])
			} else {
				tranPtrs = []*TranInfo{&tranList[i]}
			}
			cityTranMap[cityCode] = tranPtrs
		}
	}
}

// TranInfo 列车信息结构体
type TranInfo struct {
	TranNum         string    `gorm:"index:main;type:varchar(10)" json:"tranNum"`                         // 车次号
	DurationDays    uint8     `json:"durationDays"`                                                       // 一趟行程运行多少天，次日达则值应为2
	ScheduleDays    uint8     `gorm:"default:1" json:"scheduleDays"`                                      // 间隔多少天发一趟车，绝大多数是1天
	EnableLevel     uint8     `gorm:"default:1" json:"enableLevel"`                                       // 生效级别 0为最高级别，默认级别为1
	CarIds          string    `gorm:"type:varchar(100)" json:"carIds"`                                    // 车厢ID及其数量，格式如：32-1;12-2; ...
	Cars            []Car     `gorm:"-"`                                                                  //
	EnableStartDate time.Time `gorm:"type:datetime;default:'0000-00-00 00:00:00'" json:"enableStartDate"` // 生效开始日期 默认零值
	EnableEndDate   time.Time `gorm:"type:datetime;default:'9999-12-31 23:59:59'" json:"enableEndDate"`   // 生效截止日期 默认最大值
	// 关联对象
	RouteTimetable []Route                `gorm:"foreignkey:TranID" json:"routeTimetable"` // 时刻表
	SeatPriceMap   map[string]([]float32) `gorm:"-" json:"seatPriceMap"`                   // 各类席位在各路段的价格
	DBModel
}

func (t *TranInfo) getFullInfo() {
	// 获取时刻表信息
	var routes []Route
	db.Where("tran_id = ?", t.ID).Find(&routes)
	t.RouteTimetable = routes

	// 座次类型，按好到次排序
	carTypes := []string{constSeatTypeSpecial, constSeatTypeFristClass, constSeatTypeSecondClass,
		constSeatTypeAdvancedSoftSleeper, constSeatTypeSoftSleeper, constSeatTypeEMUSleeper,
		constSeatTypeMoveSleeper, constSeatTypeHardSleeper, constSeatTypeSoftSeat,
		constSeatTypeHardSeat, constSeatTypeNoSeat,
	}

	t.SeatPriceMap = make(map[string]([]float32), 0)
	var routePrices []RoutePrice
	db.Where("tran_id = ?", t.ID).Order("seat_type, route_index").Find(&routePrices)
	for cti := 0; cti < len(carTypes); cti++ {
		for ri := 0; ri < len(routePrices); ri++ {
			if routePrices[ri].SeatType == carTypes[cti] {
				prices, exist := t.SeatPriceMap[carTypes[cti]]
				if exist {
					prices = append(prices, routePrices[ri].Price)
				} else {
					prices = []float32{routePrices[ri].Price}
				}
				t.SeatPriceMap[carTypes[cti]] = prices
				break
			}
		}
	}

	// 获取车厢信息
	carSettings := strings.Split(t.CarIds, ";")
	carIds, carCounts := make([]string, len(carSettings)), make([]int, len(carSettings))
	for i := 0; i < len(carSettings); i++ {
		setting := strings.Split(carSettings[i], "-")
		if len(setting) == 2 {
			carIds[i] = setting[0]
			carCounts[i], _ = strconv.Atoi(setting[1])
		}
	}
	var cars []Car
	db.Where("id in (?)", strings.Join(carIds, ",")).Find(&cars)
	t.Cars = make([]Car, 0)
	// 保证车厢按座次类型顺序组装
	for cti := 0; cti < len(carTypes); cti++ {
		for ci := 0; ci < len(cars); ci++ {
			if cars[ci].SeatType == carTypes[cti] {
				for idx := 0; idx < len(carIds); idx++ {
					if strconv.Itoa(int(cars[ci].ID)) == carIds[idx] {
						for c := 0; c < carCounts[ci]; c++ {
							t.Cars = append(t.Cars, cars[ci])
						}
						break
					}
				}
			}
		}
	}
}

func (t *TranInfo) setRoutes(routes *[]Route) {
	t.RouteTimetable = *routes
	// 时刻表中的第一站出发日期，默认都为0001-01-01，所以终点站的日期即为其全程运行的天数
	t.DurationDays = uint8(t.RouteTimetable[len(t.RouteTimetable)-1].ArrTime.YearDay())
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

// IsMatchStation 判断当前车次在站点是否匹配
func (t *TranInfo) IsMatchStation(depS, arrS *Station) (depIdx, arrIdx uint8, ok bool) {
	depI := -1
	// 出发地与目的地是不同的城市
	if depS.CityCode != arrS.CityCode {
		// TODO: 当某车次的路线经过某城市的两个站，该怎么匹配？ 当前算法是匹配第一个，与12306逻辑一致 12306这里算是一个bug
		for i := 0; i < len(t.RouteTimetable); i++ {
			if depI == -1 && t.RouteTimetable[i].CityCode == depS.CityCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.RouteTimetable[i].CityCode == arrS.CityCode {
				// 同一城市内
				arrIdx = uint8(i)
				ok = true
				return
			}
		}
	} else { // 出发地与目的地是同一个城市
		for i := 0; i < len(t.RouteTimetable); i++ {
			if depI == -1 && t.RouteTimetable[i].StationCode == depS.StationCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.RouteTimetable[i].StationCode == arrS.StationCode {
				arrIdx = uint8(i)
				ok = true
				return
			}
		}
	}
	return
}

// // GetTranInfoAndSeatCount 返回车次信息、起止站在时刻表中的索引、各类座位余票数、可售标记、不售票说明
// func (t *TranInfo) GetTranInfoAndSeatCount(depIdx, arrIdx uint8, carTypeIdx []uint8, isStudent bool) (remain *RemainingTickets) {
// 	remain = newRemainingTickets(t, depIdx, arrIdx)
// 	// 有不售票说明，表示当前不售票，不用查余票数
// 	if t.NotSaleRemark != "" {
// 		return
// 	}
// 	// 当前时间未开售，不用查询余票数
// 	if t.SaleTicketTime.After(time.Now()) {
// 		remain.remark = t.SaleTicketTime.Format(constYMdHmFormat) + " 开售"
// 		return
// 	}
// 	seatBit := countSeatBit(depIdx, arrIdx)
// 	var noSeatCount uint8
// 	for i := 0; i < len(t.carTypesIdx)-1; i++ {
// 		start, end := t.carTypesIdx[i], t.carTypesIdx[i+1]
// 		if start == end {
// 			continue
// 		}
// 		availableSeatCount := 0
// 		for j := start; j < end && availableSeatCount < constMaxAvaliableSeatCount; j++ {
// 			for k := 0; k < len(t.Cars[j].Seats) && availableSeatCount < constMaxAvaliableSeatCount; k++ {
// 				if t.Cars[j].Seats[0].IsAvailable(seatBit, isStudent) {
// 					availableSeatCount++
// 				}
// 			}
// 			if noSeatCount < constMaxAvaliableSeatCount {
// 				noSeatCount += t.Cars[j].getAvailableNoSeatCount(seatBit, depIdx, arrIdx)
// 			}
// 		}
// 		remain.availableSeatCount[i] = availableSeatCount
// 	}
// 	remain.availableSeatCount[len(remain.availableSeatCount)-1] = int(noSeatCount)
// 	return
// }

// Route 时刻表信息
type Route struct {
	TranID          uint      // 车次ID
	TranNum         string    `gorm:"index:main;type:varchar(10)"` // 车次号
	StationIndex    uint8     // 车站索引
	StationName     string    `gorm:"type:nvarchar(20)" json:"stationName"`    // 车站名
	StationCode     string    `gorm:"type:varchar(10)"`                        // 车站编码
	CityCode        string    `gorm:"type:varchar(20)"`                        // 城市编码
	DepTime         time.Time `gorm:"type:datetime" json:"depTime"`            // 出发时间
	ArrTime         time.Time `gorm:"type:datetime" json:"arrTime"`            // 到达时间
	CheckTicketGate string    `gorm:"type:varchar(10)" json:"checkTicketGate"` // 检票口
	Platform        uint8     `json:"platform"`                                // 乘车站台
	MileageNext     float32   `json:"mileageNext"`                             // 距下一站的里程
	DBModel
}

// RoutePrice 各路段价格
type RoutePrice struct {
	TranID     uint    `gorm:"index:main"`      // 车次ID
	SeatType   string  `gorm:"type:varchar(5)"` // 座次类型
	RouteIndex uint8   // 路段索引
	Price      float32 // 价格
	DBModel
}

// Car 车厢信息结构体
type Car struct {
	TranType    string `gorm:"type:varchar(20)" json:"tranType"` // 车次类型 高铁、动车、直达等
	SeatType    string `gorm:"type:varchar(5)" json:"seatType"`  // 车厢的座位类型
	SeatCount   uint8  // 车厢内座位数(或床位数)
	NoSeatCount uint8  `json:"noSeatCount"`                     // 车厢内站票数
	Remark      string `gorm:"type:nvarchar(50)" json:"remark"` // 说明
	Seats       []Seat `gorm:"foreignkey:CarID" json:"seats"`   // 车厢的所有座位
	DBModel
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
	CarID     uint   `gorm:"index:main"`                     // 车厢ID
	SeatNum   string `gorm:"type:varchar(5)" json:"seatNum"` // 座位号
	IsStudent bool   `json:"isStudent"`                      // 是否学生票
}

func (r *Route) getStrDep() string {
	if r.DepTime.Year() == 1 {
		return constStrNullTime
	}
	return r.DepTime.Format(constHmFormat)
}

func (r *Route) getStrArr() string {
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
