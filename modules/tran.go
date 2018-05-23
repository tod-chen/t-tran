package modules

import (
	//"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"database/sql"
	// mysql
	_ "github.com/go-sql-driver/mysql"

	// mgo
	_ "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// 时间格式
	constHmFormat    = "15:04"            // 时间格式 小时:分钟
	constYmdFormat   = "2006-01-02"       // 时间格式 年-月-日
	constYMdHmFormat = "2006-01-02 15:04" // 时间格式 年-月-日 小时:分钟

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
	constSeatTypeMoveSleeper         = "MS"  // 动卧
	constSeatTypeHardSleeper         = "HS"  // 硬卧
	constSeatTypeSoftSeat            = "SST" // 软座
	constSeatTypeHardSeat            = "HST" // 硬座
	constSeatTypeNoSeat              = "NST" // 无座
)

var (
	// 所有车次信息，不参与订票，用于查询列车的时刻表和各路段各座次的价格
	tranList []Tran
	// 各城市与经过该城市的列车映射
	cityTranMap map[string]([]*Tran)
	// 列车按日期安排表
	tranScheduleMap map[string]([]Tran)
)

func initTran() {
	initTranList()
	initCityTranMap()
	initTranScheduleMap()
}

func initTranList() {
	fmt.Println("begin init tranList")
	defer fmt.Println("end init tranList")

	session := getMgoSession()
	defer session.Close()
	db := session.DB(mgoDbName)

	cfg := db.C("config")
	tranCfg := new(config)
	err := cfg.Find(bson.M{"key": "transInfoSavedInMongo"}).One(&tranCfg)
	if err != nil {
		panic(err)
	}
	if tranCfg.Value == "0" {
		initTranListFromMySQL()
		tranCfg.Value = "1"
		cfg.Update(bson.M{"key": "transInfoSavedInMongo"}, tranCfg)
	} else {
		trans := session.DB(mgoDbName).C("tranInfo")
		err := trans.Find(bson.M{}).All(&tranList)
		if err != nil {
			panic("query error")
		}
	}
}

func initTranListFromMySQL() {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/t-tran?charset=utf8")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	query := "SELECT DISTINCT tranNum FROM routes ORDER BY tranNum"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		t := Tran{ScheduleIntervalDays: 1}
		err = rows.Scan(&t.TranNum)
		if err != nil {
			panic(err)
		}
		routes := make([]Route, 0, 10) // 根据车次号从路线表中读取
		queryRoute := "select stationName, stationCode, cityCode, depTime, arrTime from routes where tranNum = '" + t.TranNum + "' order by stationIndex"
		routeRows, _ := db.Query(queryRoute)
		defer routeRows.Close()
		routeIdx, arrTime, depTime, date := 0, "", "", time.Now()
		for routeRows.Next() {
			r := new(Route)
			routeRows.Scan(&r.StationName, &r.StationCode, &r.CityCode, &depTime, &arrTime)
			// 设置出发时间和到达时间，这里有个默认规则：列车在两个站之间的行驶时间不超过24小时，
			//  如果某列车不满足此规则，则下面这段代码得调整逻辑
			if arrTime != "" {
				r.ArrTime, err = time.Parse(time.RFC3339, date.Format(constYmdFormat)+"T"+arrTime+"Z")
				if err != nil {
					panic(err)
				}
				// 到达时间早于上一站的出发时间，则表示跨天了
				if routeIdx > 0 && r.ArrTime.Before(routes[routeIdx-1].DepTime) {
					r.ArrTime = r.ArrTime.Add(constOneDayDuration)
					date = date.Add(constOneDayDuration)
				}
			}
			if depTime != "" {
				r.DepTime, err = time.Parse(time.RFC3339, date.Format(constYmdFormat)+"T"+depTime+"Z")
				if err != nil {
					panic(err)
				}
				// 出发时间早于上一站到达时间，则表示跨天了
				if routeIdx > 0 && r.DepTime.Before(r.ArrTime) {
					r.DepTime = r.DepTime.Add(constOneDayDuration)
					date = date.Add(constOneDayDuration)
				}
			}
			routes = append(routes, *r)
			routeIdx++
		}
		t.setRoutes(&routes)
		tranList = append(tranList, t)
	}
	session := getMgoSession()
	defer session.Close()
	trans := session.DB(mgoDbName).C("tranInfo")
	for i := 0; i < len(tranList); i++ {
		err = trans.Insert(tranList[i])
		if err != nil {
			panic(err)
		}
	}
}

func initCityTranMap() {
	fmt.Println("begin init cityTranMap")
	defer fmt.Println("end init cityTranMap")
	cityTranMap = make(map[string]([]*Tran), constCityCount)
	for i := 0; i < len(tranList); i++ {
		for j := 0; j < len(tranList[i].RouteTimetable); j++ {
			cityCode := tranList[i].RouteTimetable[j].CityCode
			tranPtrs, exist := cityTranMap[cityCode]
			if exist {
				tranPtrs = append(tranPtrs, &tranList[i])
			} else {
				tranPtrs = []*Tran{&tranList[i]}
			}
			cityTranMap[cityCode] = tranPtrs
		}
	}
}

// 初始化列车排班，后期此部分数据会存入数据库
func initTranScheduleMap() {
	fmt.Println("begin init tranScheduleMap")
	defer fmt.Println("end init tranScheduleMap")

	tranScheduleMap = make(map[string]([]Tran), 30)
	session := getMgoSession()
	defer session.Close()
	schedules := session.DB(mgoDbName).C("tranSchedules")
	now := time.Now()
	for d := 0; d <= constDays; d++ {
		date := now.AddDate(0, 0, d).Format(constYmdFormat)
		trans := make([]Tran, 0, constTranCount)
		if err := schedules.Find(bson.M{"departureDate": date}).All(&trans); err != nil {
			panic(err)
		}
		tranScheduleMap[date] = trans
	}
}

// 余票信息按出发时间排序，也可以丢到前端去排序，js的排序还是比较方便的
type remainingTicketSort []*RemainingTickets

func (r remainingTicketSort) Len() int           { return len(r) }
func (r remainingTicketSort) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r remainingTicketSort) Less(i, j int) bool { return r[i].depTime.Before(r[j].depTime) }

// RemainingTickets 余票信息结构
type RemainingTickets struct {
	tranNum            string    // 车次号
	date               string    // 发车日期
	routeCount         uint8     // 列车运行的路段数
	depIdx             uint8     // 出发站索引，值为0时表示出发站为起点站，否则表示路过
	depStationCode     string    // 出发站编码
	depTime            time.Time // 出发时间, 满足条件的列车需根据出发时间排序
	arrIdx             uint8     // 目的站索引，值与routeCount相等时表示目的站为终点，否则表示路过
	arrStationCode     string    // 目的站编码
	arrTime            string    // 到达时间
	costTime           string    // 历时，根据出发时间与历时可计算出跨天数，在前端计算即可
	availableSeatCount []int     // 各座次余票数
	remark             string    // 不售票的说明
}

func newRemainingTickets(t *Tran, depIdx, arrIdx uint8) *RemainingTickets {
	r := &RemainingTickets{
		tranNum:        t.TranNum,
		date:           t.RouteTimetable[0].DepTime.Format(constYmdFormat),
		routeCount:     uint8(len(t.RouteTimetable)) - 1,
		depIdx:         depIdx,
		depStationCode: t.RouteTimetable[depIdx].StationCode,
		depTime:        t.RouteTimetable[depIdx].DepTime,
		arrIdx:         arrIdx,
		arrStationCode: t.RouteTimetable[arrIdx].StationCode,
		arrTime:        t.RouteTimetable[arrIdx].ArrTime.Format(constHmFormat),
		costTime:       t.RouteTimetable[arrIdx].ArrTime.Sub(t.RouteTimetable[depIdx].DepTime).String(),
		// 初始化全部没有此类
		availableSeatCount: []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		remark:             t.NotSaleRemark,
	}
	return r
}

// 结果转为字符串
func (r *RemainingTickets) toString() string {
	list := []string{r.tranNum, r.date, strconv.Itoa(int(r.routeCount)),
		strconv.Itoa(int(r.depIdx)), r.depStationCode, r.depTime.Format(constHmFormat),
		strconv.Itoa(int(r.arrIdx)), r.arrStationCode, r.arrTime, r.costTime}
	countList := make([]string, 12)
	count := 0
	for i := 0; i < len(r.availableSeatCount); i++ {
		count = r.availableSeatCount[i]
		if count == 0 {
			countList[i] = "无"
		}
		if 0 < count && count < constMaxAvaliableSeatCount {
			countList[i] = strconv.Itoa(count)
		}
		if count >= constMaxAvaliableSeatCount {
			countList[i] = "有"
		}
	}
	countList[11] = r.remark
	list = append(list, countList...)
	return strings.Join(list, "|")
}

// QueryMatchTransInfo 获取所选出发日期中， 经过出发站、目的站的车次及其各类余票数量
func QueryMatchTransInfo(depStationName, arrStationName string, queryDate time.Time, isStudent bool) (result []string) {
	depS := getStationInfoByName(depStationName)
	depTrans, exist := cityTranMap[depS.CityCode]
	if !exist {
		return
	}
	arrS := getStationInfoByName(arrStationName)
	availableTime := time.Now().Add(constQueryTranDelay * time.Minute)
	resultCh := make(chan *RemainingTickets, 20)
	defer close(resultCh)
	matchTranCount := 0
	for i := 0; i < len(depTrans); i++ {
		depIdx, arrIdx, ok := depTrans[i].IsMatchStation(depS, arrS)
		if !ok {
			continue
		}
		// 根据起点站和出发站之间的时间跨度，以及查询日期，算出该列车的发车日期
		firstStationDepDate := depTrans[i].RouteTimetable[0].DepTime
		depStationDate := depTrans[i].RouteTimetable[depIdx].DepTime
		subDays := int(depStationDate.Sub(firstStationDepDate).Hours() / 24)
		dateKey := queryDate.AddDate(0, 0, -subDays).Format(constYmdFormat)
		// 遍历当前列车出发日的所有列车，找出匹配项，这里需要做优化，改用二分法查找(需要对tranList排序)
		trans := tranScheduleMap[dateKey]
		for i := 0; i < len(trans); i++ {
			if depTrans[i].TranNum != trans[i].TranNum {
				continue
			}
			depTime := trans[i].RouteTimetable[depIdx].DepTime
			// 发车时间在所查询的日期内, 跳出外层循环
			if queryDate.Before(depTime) && depTime.Before(queryDate.Add(constOneDayDuration)) {
				ok = false
				// 发车前20分钟内，不予查询
				if !depTime.Before(availableTime) {
					break
				}
				matchTranCount++
				goPool.Take()
				go func(tranInfo *Tran, depIdx, arrIdx uint8) {
					defer goPool.Return()
					resultCh <- trans[i].GetTranInfoAndSeatCount(depIdx, arrIdx, tranInfo.CarTypesIdx, isStudent)
				}(depTrans[i], depIdx, arrIdx)
				break
			}
		}
	}
	resultList := make([]*RemainingTickets, matchTranCount)
	for i := 0; i < matchTranCount; i++ {
		resultList[i] = <-resultCh
	}
	sort.Sort(remainingTicketSort(resultList))
	result = make([]string, matchTranCount)
	for i := 0; i < len(resultList); i++ {
		result[i] = resultList[i].toString()
	}
	return
}

// QueryRouteResult 查询时刻表的结果
type QueryRouteResult struct {
	name     string // 车站名
	depTime  string // 出发时间
	arrTime  string // 到达时间
	stayTime string // 停留时间
}

// QueryRoutetable 查询时刻表
func QueryRoutetable(tranNum string) (result []QueryRouteResult) {
	for i := 0; i < len(tranList); i++ {
		if tranList[i].TranNum == tranNum {
			routeLen := len(tranList[i].RouteTimetable)
			result = make([]QueryRouteResult, routeLen)
			for j := 0; j < routeLen; j++ {
				routeInfo := QueryRouteResult{
					name:     tranList[i].RouteTimetable[j].StationName,
					depTime:  tranList[i].RouteTimetable[j].getStrDep(),
					arrTime:  tranList[i].RouteTimetable[j].getStrArr(),
					stayTime: tranList[i].RouteTimetable[j].getStrStayTime(),
				}
				result = append(result, routeInfo)
			}
			break
		}
	}
	return
}

// QuerySeatPrice 查询票价
func QuerySeatPrice(tranNum string, depIdx, arrIdx uint8) (result map[string]float32) {
	for i := 0; ; i++ {
		if tranList[i].TranNum == tranNum {
			result = tranList[i].getSeatPrice(depIdx, arrIdx)
			break
		}
	}
	return
}

//////////////////////////////////////////////
///          列车结构体及其方法              ///
//////////////////////////////////////////////

// Tran 列车信息结构体
type Tran struct {
	TranNum        string    `bson:"tranNum"`        // 车次号
	DepartureDate  string    `bson:"departureDate"`  // 发车日期
	SaleTicketTime time.Time `bson:"saleTicketTime"` // 售票时间
	NotSaleRemark  string    `bson:"notSaleRemark"`  // 不售票的说明，路线调整啥的
	RouteTimetable []Route   `bson:"routeTimetable"` // 时刻表
	Cars           []Car     `bson:"cars"`           // 车厢
	// 各种类型车厢的起始索引
	// 依次是：商务座、一等座、二等座、高级软卧、软卧、动卧、硬卧、软座、硬座
	// 若值是[0, 1, 4, 6, 6, 6, 6, 6, 10, 15]，
	// 表示第一个车厢是商务座; 第2-4个车厢是一等座; 第5-6个车厢是二等座; 高级软卧、软卧、动卧、硬卧都没有; 第7-10个是软座; 第11-15个是硬座;
	CarTypesIdx          []uint8   `bson:"carTypesIdx"`
	FullSeatBit          uint64    `bson:"fullSeatBit"`          // 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
	RunDays              uint8     `bson:"runDays"`              // 一趟行程运行多少天，次日达则值应为2
	ScheduleIntervalDays uint8     `bson:"scheduleIntervalDays"` // 间隔多少天发一趟车，绝大多数是1天
	EnableLevel          uint8     `bson:"enableLevel"`          // 生效级别 0为最高级别，默认级别为1
	EnableStartDate      time.Time `bson:"enableStartDate"`      // 时刻表生效截止日期
	EnableEndDate        time.Time `bson:"enableEndDate"`        // 车厢信息生效截止日期
	// 各类席位在各路段的价格
	SeatPriceMap map[string]([]float32) `bson:"seatPriceMap"`
}

// QueryTrans 查询列车信息
func QueryTrans(tranNum, tranType string, pageIdx, pageSize int) (result []Tran) {
	session := getMgoSession()
	defer session.Close()
	tranCol := session.DB(mgoDbName).C("tranInfo")
	err := tranCol.Find(bson.M{"tranNum": bson.M{"$regex": tranNum, "$options": "$i"}}).Sort("tranNum").Skip((pageIdx - 1) * pageSize).Limit(pageSize).All(&result)
	if err != nil {
		panic(err)
	}
	return
}

func (t *Tran) setRoutes(routes *[]Route) {
	t.RouteTimetable = *routes
	routeCount := len(t.RouteTimetable) - 1
	t.FullSeatBit = countSeatBit(0, uint8(routeCount))
	depTime := t.RouteTimetable[0].DepTime
	arrTime := t.RouteTimetable[routeCount].ArrTime
	t.RunDays = uint8(arrTime.YearDay() - depTime.YearDay() + 1)
}

// 根据起止站获取各类座位的票价
func (t *Tran) getSeatPrice(depIdx, arrIdx uint8) (result map[string]float32) {
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
func (t *Tran) IsMatchStation(depS, arrS *station) (depIdx, arrIdx uint8, ok bool) {
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

// GetTranInfoAndSeatCount 返回车次信息、起止站在时刻表中的索引、各类座位余票数、可售标记、不售票说明
func (t *Tran) GetTranInfoAndSeatCount(depIdx, arrIdx uint8, carTypeIdx []uint8, isStudent bool) (remain *RemainingTickets) {
	remain = newRemainingTickets(t, depIdx, arrIdx)
	// 有不售票说明，表示当前不售票，不用查余票数
	if t.NotSaleRemark != "" {
		return
	}
	// 当前时间未开售，不用查询余票数
	if t.SaleTicketTime.After(time.Now()) {
		remain.remark = t.SaleTicketTime.Format(constYMdHmFormat) + " 开售"
		return
	}
	seatBit := countSeatBit(depIdx, arrIdx)
	var noSeatCount uint8
	for i := 0; i < len(t.CarTypesIdx)-1; i++ {
		start, end := t.CarTypesIdx[i], t.CarTypesIdx[i+1]
		if start == end {
			continue
		}
		availableSeatCount := 0
		for j := start; j < end && availableSeatCount < constMaxAvaliableSeatCount; j++ {
			for k := 0; k < len(t.Cars[j].Seats) && availableSeatCount < constMaxAvaliableSeatCount; k++ {
				if t.Cars[j].Seats[0].IsAvailable(seatBit, isStudent) {
					availableSeatCount++
				}
			}
			if noSeatCount < constMaxAvaliableSeatCount {
				noSeatCount += t.Cars[j].getAvailableNoSeatCount(seatBit, depIdx, arrIdx)
			}
		}
		remain.availableSeatCount[i] = availableSeatCount
	}
	remain.availableSeatCount[len(remain.availableSeatCount)-1] = int(noSeatCount)
	return
}

//////////////////////////////////////////////
///          车厢结构体及其对应方法          ///
//////////////////////////////////////////////

// Car 车厢信息结构体
type Car struct {
	CarNum                 uint8   `bson:"carNum"`                 // 车厢编号
	SeatType               string  `bson:"seatType"`               // 车厢的座位类型
	Seats                  []Seat  `bson:"seats"`                  // 车厢的所有座位
	NoSeatCount            uint8   `bson:"noSeatCount"`            // 车厢内站票数
	EachRouteTravelerCount []uint8 `bson:"eachRouteTravelerCount"` // 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	sync.RWMutex                   // 读写锁，用于保护各路段乘客人数字段
}

// getAvailableSeat 获取可预订的座位,是否获取成功标记,是否为拼凑的站票标记
func (c *Car) getAvailableSeat(seatBit uint64, isStudent bool) (s *Seat, ok bool) {
	for i := 0; i < len(c.Seats); i++ {
		if c.Seats[i].IsAvailable(seatBit, isStudent) {
			return &c.Seats[i], true
		}
	}
	return nil, false
}

func (c *Car) getAvailableNoSeat(seatBit uint64, depIdx, arrIdx uint8) (s *Seat, ok bool) {
	if c.NoSeatCount == 0 {
		return nil, false
	}
	// 下面开始查找拼凑的站票
	// 非站票数与站票数之和
	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	defer c.RUnlock()
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	if totalSeatCount-int(maxTravelerCountInRoute) > 0 {
		s = &Seat{SeatNum: "", SeatBit: seatBit}
		return s, true
	}
	return nil, false
}

func (c *Car) getAvailableNoSeatCount(seatBit uint64, depIdx, arrIdx uint8) uint8 {
	// 车厢未设置站票时，直接返回 0
	if c.NoSeatCount == 0 {
		return 0
	}
	// 非站票数与站票数之和
	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	defer c.RUnlock()
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	return uint8(totalSeatCount) - maxTravelerCountInRoute
}

// 某座位被占用
func (c *Car) occupySeat(depIdx, arrIdx uint8) bool {
	if c.NoSeatCount == 0 {
		return true
	}
	c.Lock()
	defer c.Unlock()
	var maxCount uint8
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxCount {
			maxCount = c.EachRouteTravelerCount[i]
		}
	}
	if maxCount >= uint8(len(c.Seats))+c.NoSeatCount {
		return false
	}
	for i := depIdx; i < arrIdx; i++ {
		c.EachRouteTravelerCount[i]++
	}
	return true
}

// 某座位被释放
func (c *Car) releaseSeat(depIdx, arrIdx uint8) {
	if c.NoSeatCount != 0 {
		c.Lock()
		defer c.Unlock()
		for i := depIdx; i < arrIdx; i++ {
			c.EachRouteTravelerCount[i]--
		}
	}
}

//////////////////////////////////////////////
///          座位结构体及其对应方法          ///
//////////////////////////////////////////////

// Seat 座位信息结构体
type Seat struct {
	SeatNum    string `bson:"seatNum"`   // 座位号
	IsStudent  bool   `bson:"isStudent"` // 是否学生票
	SeatBit    uint64 `bson:"seatBit"`   // 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	sync.Mutex        // 锁，订票与退票均需要锁
}

// IsAvailable 根据路段和乘客类型判断能否订票
func (s *Seat) IsAvailable(seatBit uint64, isStudent bool) bool {
	return (s.IsStudent == isStudent) && (s.SeatBit^seatBit == s.SeatBit+seatBit)
}

// Book 订票
func (s *Seat) Book(seatBit, tranFullSeatBit uint64, isStudent bool) bool {
	s.Lock()
	defer s.Unlock()
	if !s.IsAvailable(seatBit, isStudent) {
		return false
	}
	s.SeatBit ^= seatBit
	return true
}

// Release 退票或取消订单，释放座位对应路段的资源
func (s *Seat) Release(seatBit uint64) {
	s.Lock()
	defer s.Unlock()
	s.SeatBit ^= (^seatBit)
}

func countSeatBit(depIdx, arrIdx uint8) (result uint64) {
	for i := depIdx; i <= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}
