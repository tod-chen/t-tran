package modules

import (
	"sync"
	"strings"
	"strconv"
	"time"
	"fmt"
	"database/sql"
	"sort"
)

const (
	// 时间格式 小时:分钟 
	constHmFormat = "15:04"
	// 时间格式 年-月-日
	constYmdFormat = "2006-01-02"
	// 时间格式 年-月-日 小时:分钟
	constYMdHmFormat = constYmdFormat + " " + constHmFormat
	// 一天的长度
	constOneDayDuration = 24 * time.Hour
	// 车次数量
	constTranCount = 13000
	// 可提前订票天数
	constDays = 30
	// 有火车经过的城市数量
	constCityCount = 620
	// 查询时距离当前时间多长时间内发车的车次不予显示 单位：分钟
	constQueryTranDelay = 20
	// 未完成订单有效时间 单位：分钟
	constUnpayOrderAvaliableTime = 45
	// 查询余票数量的最大值 超过此值时，显示“有”
	constMaxAvaliableSeatCount = 100
	// 座位类型数
	constSeatTypeCount = 10

	// 各类座位的名称
	constSeatTypeSpecial = "S" // 商务座
	constSeatTypeFristClass = "FC"	// 一等座
	constSeatTypeSecondClass = "SC"	// 二等座
	constSeatTypeAdvancedSoftSleeper = "ASS"	// 高级软卧
	constSeatTypeSoftSleeper = "SS" // 软卧
	constSeatTypeMoveSleeper = "MS"	// 动卧
	constSeatTypeHardSleeper = "HS"	// 硬卧
	constSeatTypeSoftSeat = "SST"	// 软座
	constSeatTypeHardSeat = "HST"	// 硬座
	constSeatTypeNoSeat = "NST"		// 无座
)

var (
	// 所有车次信息，不参与订票，用于查询列车的时刻表和各路段各座次的价格
	tranList []tran
	// 各城市与经过该城市的列车映射
	cityTranMap map[string]([]*tran)
	// 列车按日期安排表
	tranScheduleMap map[string]([]tran)
)

func initTran(db *sql.DB){
	initTranList(db)
	initCityTranMap()
	initTranScheduleMap()
}

func initTranList(db *sql.DB){
	fmt.Println("begin init tranList")
	defer fmt.Println("end init tranList")
	tranList = make([]tran, 0, constTranCount)
	query := "select tranNum, depTime, runDays from traninfo"
	rows, err := db.Query(query)
	if err != nil {
		panic("query error")
	}
	defer rows.Close()
	var (
		tranNum string
		depTime string
		runDays int
	)
	for rows.Next() {
		err = rows.Scan(&tranNum, &depTime, &runDays)
		if err != nil {
			panic(err)
		}
		t := tran{}
		t.tranNum = tranNum
		t.runDays = uint8(runDays)
		tranList = append(tranList, t)
	}
	var (
		arrTime string
		routeIdx int
	)
	for i:=0; i<len(tranList); i++ {
		// 初始化所有列车的组合字段
		routes := make([]route, 0, 10)	// 根据车次号从路线表中读取
		queryRoute := "select stationName, stationCode, cityCode, arrTime, depTime from routes where tranNum = '" + tranList[i].tranNum + "' order by stationIndex"
		routeRows, _ := db.Query(queryRoute)		
		defer routeRows.Close()
		routeIdx = 0
		date := time.Now()
		for routeRows.Next() {
			r := new(route)
			routeRows.Scan(&r.stationName, &r.stationCode, &r.cityCode, &arrTime, &depTime)
			// 设置出发时间和到达时间，这里有个默认规则：列车在两个站之间的行驶时间不超过24小时，
			//  如果某列车不满足此规则，则下面这段代码得调整逻辑
			if arrTime != "" {
				r.arrTime, err = time.Parse(time.RFC3339, date.Format(constYmdFormat) + "T" + arrTime + "Z")
				if err != nil {
					panic(err)
				}
				// 到达时间早于上一站的出发时间，则表示跨天了
				if routeIdx > 0 && r.arrTime.Before(routes[routeIdx-1].depTime) {
					r.arrTime = r.arrTime.Add(constOneDayDuration)
					date = date.Add(constOneDayDuration)
				}
			}
			if depTime != "" {
				r.depTime, err = time.Parse(time.RFC3339, date.Format(constYmdFormat) + "T" + depTime + "Z")
				if err != nil {
					panic(err)
				}
				// 出发时间早于上一站到达时间，则表示跨天了
				if r.depTime.Before(r.arrTime) {
					r.depTime = r.depTime.Add(constOneDayDuration)
					date = date.Add(constOneDayDuration)
				}
			}
			routes = append(routes, *r)
			routeIdx++
		}
		routeLen := len(routes)
		tranList[i].fullSeatBit = countSeatBit(0, uint8(routeLen - 1))
		tranList[i].routeTimetable = routes
		// 根据车次号首字母设置其各座次的车厢起止索引 TODO：此信息应该放在数据库中，以便针对性修改
		switch string([]byte(tranList[i].tranNum)[:1]) {
		case "G": // 高铁
			tranList[i].carTypesIdx = []uint8{0, 1, 4, 14, 14, 14, 14, 14, 14, 14}
		case "C": // 城际
			tranList[i].carTypesIdx = []uint8{0, 0, 2, 7, 7, 7, 7, 7, 7, 7}
		case "D": // 动车
			tranList[i].carTypesIdx = []uint8{0, 0, 2, 7, 7, 7, 7, 7, 7, 7}
		case "T": // 特快
			tranList[i].carTypesIdx = []uint8{0, 0, 0, 0, 0, 3, 3, 9, 9, 15}
		case "Z": // 直达
			tranList[i].carTypesIdx = []uint8{0, 0, 0, 0, 0, 3, 3, 15, 15, 15}
		case "K": // 普快
			tranList[i].carTypesIdx = []uint8{0, 0, 0, 0, 0, 3, 3, 9, 9, 15}
		default:
			tranList[i].carTypesIdx = []uint8{0, 0, 0, 0, 0, 3, 3, 9, 9, 15}
		}
	}
}

func initCityTranMap(){
	fmt.Println("begin init cityTranMap")
	defer fmt.Println("end init cityTranMap")
	cityTranMap = make(map[string]([]*tran), constCityCount)
	for i:=0; i<len(tranList); i++ {
		for j:=0; j<len(tranList[i].routeTimetable); j++ {
			cityCode := tranList[i].routeTimetable[j].cityCode
			tranPtrs, exist := cityTranMap[cityCode]
			if exist {
				tranPtrs = append(tranPtrs, &tranList[i])
			} else{
				tranPtrs = []*tran { &tranList[i] }
			}
			cityTranMap[cityCode] = tranPtrs
		}
	}
}

// 初始化列车排班，后期此部分数据会存入数据库
func initTranScheduleMap(){
	fmt.Println("begin init tranScheduleMap")
	defer fmt.Println("end init tranScheduleMap")
	tranScheduleMap = make(map[string]([]tran), constDays)
	now := time.Now()
	var zeroTime time.Time
	for i:=0; i<len(tranList); i++ {
		for d:=0; d<constDays; d++ {
			date := now.AddDate(0, 0, d).Format(constYmdFormat)
			trans, exist := tranScheduleMap[date]
			if !exist {
				trans = make([]tran, 0, constTranCount)
			}
			tmepTran := tranList[i].copyForSchedule()
			tranList[i].departureDate = now.AddDate(0, 0, d).Format(constYmdFormat)
			for r:=0; r<len(tmepTran.routeTimetable);r++ {
				if tmepTran.routeTimetable[r].depTime != zeroTime {
					tmepTran.routeTimetable[r].depTime = tmepTran.routeTimetable[r].depTime.AddDate(0, 0, d)
				}
				if tmepTran.routeTimetable[r].arrTime != zeroTime {
					tmepTran.routeTimetable[r].arrTime = tmepTran.routeTimetable[r].arrTime.AddDate(0, 0, d)
				}
			}
			tranScheduleMap[date] = append(trans, tranList[i])
		}
	}
}

// 余票信息按出发时间排序，也可以丢到前端去排序，js的排序还是比较方便的
type remainingTicketSort []*remainingTickets
func (r remainingTicketSort) Len() int { return len(r) }
func (r remainingTicketSort) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r remainingTicketSort) Less(i, j int) bool { return r[i].depTime.Before(r[j].depTime) }

// 余票信息结构
type remainingTickets struct{
	tranNum string	// 车次号
	date string		// 发车日期
	routeCount uint8 // 列车运行的路段数
	depIdx uint8	// 出发站索引，值为0时表示出发站为起点站，否则表示路过
	depStationCode string // 出发站编码
	depTime time.Time	// 出发时间, 满足条件的列车需根据出发时间排序
	arrIdx uint8	// 目的站索引，值与routeCount相等时表示目的站为终点，否则表示路过
	arrStationCode string // 目的站编码
	arrTime string	// 到达时间
	costTime string	// 历时，根据出发时间与历时可计算出跨天数，在前端计算即可
	availableSeatCount []int	// 各座次余票数
	remark string // 不售票的说明
}

func newRemainingTickets(t *tran, depIdx, arrIdx uint8) *remainingTickets {
	r := &remainingTickets {
		tranNum: t.tranNum,
		date: t.routeTimetable[0].depTime.Format(constYmdFormat),
		routeCount: uint8(len(t.routeTimetable)) - 1,
		depIdx: depIdx,
		depStationCode: t.routeTimetable[depIdx].stationCode,
		depTime: t.routeTimetable[depIdx].depTime,
		arrIdx: arrIdx,
		arrStationCode: t.routeTimetable[arrIdx].stationCode,
		arrTime: t.routeTimetable[arrIdx].arrTime.Format(constHmFormat),
		costTime: t.routeTimetable[arrIdx].arrTime.Sub(t.routeTimetable[depIdx].depTime).String(),
		// 初始化全部没有此类
		availableSeatCount: []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		remark: t.notSaleRemark,
	}
	return r
}

// 结果转为字符串
func (r *remainingTickets) toString() string{
	list := []string{ r.tranNum, r.date, strconv.Itoa(int(r.routeCount)), 
		strconv.Itoa(int(r.depIdx)), r.depStationCode, r.depTime.Format(constHmFormat),
		strconv.Itoa(int(r.arrIdx)), r.arrStationCode, r.arrTime, r.costTime }
	countList := make([]string, 12)
	count := 0 
	for i:=0; i<len(r.availableSeatCount); i++ {
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
func QueryMatchTransInfo(depStationCode, arrStationCode string, queryDate time.Time, isAdult bool)(result []string) {
	depCityCode := getCityCodeByStationCode(depStationCode)
	depTrans, exist := cityTranMap[depCityCode]
	if !exist {
		return
	}
	arrCityCode := getCityCodeByStationCode(arrStationCode)
	availableTime := time.Now().Add(constQueryTranDelay * time.Minute)
	resultCh := make(chan *remainingTickets, 20)
	matchTranCount := 0
	for i:=0; i<len(depTrans); i++ {
		depIdx, arrIdx, ok := depTrans[i].IsMatchStation(depCityCode, depStationCode, arrCityCode, arrStationCode)
		if !ok {
			continue
		}
		// 根据起点站和出发站之间的时间跨度，以及查询日期，算出该列车的发车日期
		firstStationDepDate := depTrans[i].routeTimetable[0].depTime
		depStationDate := depTrans[i].routeTimetable[depIdx].depTime
		subDays := int(depStationDate.Sub(firstStationDepDate).Hours() / 24)
		dateKey := queryDate.AddDate(0, 0, -subDays).Format(constYmdFormat)
		// 遍历当前列车出发日的所有列车，找出匹配项，这里需要做优化，改用二分法查找(需要对tranList排序)
		trans := tranScheduleMap[dateKey]
		for i:=0; i<len(trans); i++ {
			if depTrans[i].tranNum != trans[i].tranNum {
				continue
			}
			depTime := trans[i].routeTimetable[depIdx].depTime
			// 发车时间在所查询的日期内, 跳出外层循环
			if queryDate.Before(depTime) && depTime.Before(queryDate.Add(constOneDayDuration)) {
				ok = false
				// 发车前20分钟内，不予查询
				if !depTime.Before(availableTime) {
					break
				}
				matchTranCount++
				goPool.Take()
				go func(tranInfo *tran, depIdx, arrIdx uint8){
					defer goPool.Return()
					resultCh <- trans[i].GetTranInfoAndSeatCount(depIdx, arrIdx, tranInfo.carTypesIdx, isAdult)
				}(depTrans[i], depIdx, arrIdx)					
				break
			}
		}
	}
	resultList := make([]*remainingTickets, matchTranCount)
	for i:=0; i<matchTranCount; i++ {
		resultList[i] = <- resultCh
	}
	sort.Sort(remainingTicketSort(resultList))
	result = make([]string, matchTranCount)
	for i:=0; i<len(resultList); i++ {
		result[i] = resultList[i].toString()
	}
	return
}

// QueryRouteResult 查询时刻表的结果
type QueryRouteResult struct {
	name string
	depTime string
	arrTime string
	stayTime string
}

// QueryRoutetable 查询时刻表
func QueryRoutetable(tranNum string)(result []QueryRouteResult){
	for i:=0; i<len(tranList); i++ {
		if tranList[i].tranNum == tranNum {
			routeLen := len(tranList[i].routeTimetable)
			result = make([]QueryRouteResult, routeLen)
			for j:=0; j<routeLen; j++ {
				routeInfo := QueryRouteResult{
					name: tranList[i].routeTimetable[j].stationName,
					depTime: tranList[i].routeTimetable[j].getStrDep(),
					arrTime: tranList[i].routeTimetable[j].getStrArr(),
					stayTime: tranList[i].routeTimetable[j].getStrStayTime(),
				}
				result = append(result, routeInfo)
			}
			break
		}
	}
	return
}

// QuerySeatPrice 查询票价
func QuerySeatPrice(tranNum string, depIdx, arrIdx uint8)(result map[string]float32){
	for i:=0; ; i++ {
		if tranList[i].tranNum == tranNum {
			result = tranList[i].getSeatPrice(depIdx, arrIdx)
			break
		}
	}
	return
}

//////////////////////////////////////////////
///          列车结构体及其方法              ///
//////////////////////////////////////////////
type tran struct{
	id int // 车次ID
	tranNum string // 车次号
	departureDate string // 发车日期
	saleTicketTime time.Time // 售票时间
	notSaleRemark string // 不售票的说明，路线调整啥的
	routeTimetable []route // 时刻表
	seatPriceMap map[string]([]float32) // 各类席位在各路段的价格
	cars []car // 车厢
	// 各种类型车厢的起始索引
	// 依次是：商务座、一等座、二等座、高级软卧、软卧、动卧、硬卧、软座、硬座
	// 若值是[0, 1, 4, 6, 6, 6, 6, 6, 10, 15]，
	// 表示第一个车厢是商务座; 第2-4个车厢是一等座; 第5-6个车厢是二等座; 高级软卧、软卧、动卧、硬卧都没有; 第7-10个是软座; 第11-15个是硬座;
	carTypesIdx []uint8
	// 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
	fullSeatBit uint64
	// 一趟行程运行多少天，次日达则值应为2
	runDays uint8
}

func (t *tran)copyForSchedule() tran {
	m := tran{
		tranNum: t.tranNum,
		saleTicketTime: t.saleTicketTime,
		notSaleRemark: t.notSaleRemark,
		routeTimetable: t.routeTimetable[:],
		fullSeatBit: t.fullSeatBit,
	}
	return m
}

// 根据起止站获取各类座位的票价
func (t *tran)getSeatPrice(depIdx, arrIdx uint8)(result map[string]float32){
	for seatType, eachRoutePrice := range t.seatPriceMap {
		var price float32
		for i:=depIdx; i<arrIdx; i++ {
			price += eachRoutePrice[i]
		}
		result[seatType] = price
	}
	return
}

// IsMatchStation 判断当前车次在站点是否匹配
func (t *tran)IsMatchStation(depCityCode, depStationCode, arrCityCode, arrStationCode string) (depIdx, arrIdx uint8, ok bool){
	depI := -1
	// 出发地与目的地是不同的城市
	if depCityCode != arrCityCode {
		// TODO: 当某车次的路线经过某城市的两个站，该怎么匹配？ 当前算法是匹配第一个，与12306逻辑一致 12306这里算是一个bug	
		for i:=0; i<len(t.routeTimetable); i++ {
			if depI == -1 && t.routeTimetable[i].cityCode == depCityCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.routeTimetable[i].cityCode == arrCityCode {
				// 同一城市内
				arrIdx = uint8(i)
				ok = true
				return
			}
		}
	} else { // 出发地与目的地是同一个城市
		for i:=0; i<len(t.routeTimetable); i++ {
			if depI == -1 && t.routeTimetable[i].stationCode == depStationCode {
				depI = i
				depIdx = uint8(i)
			}
			if depI != -1 && t.routeTimetable[i].stationCode == arrStationCode {
				arrIdx = uint8(i)
				ok = true
				return
			}
		}
	}
	return
}

// GetTranInfoAndSeatCount 返回车次信息、起止站在时刻表中的索引、各类座位余票数、可售标记、不售票说明
func (t *tran)GetTranInfoAndSeatCount(depIdx, arrIdx uint8, carTypeIdx []uint8, isAdult bool)(remain *remainingTickets) {
	remain = newRemainingTickets(t, depIdx, arrIdx)
	// 有不售票说明，表示当前不售票，不用查余票数
	if t.notSaleRemark != "" {
		return
	}
	// 当前时间未开售，不用查询余票数
	if t.saleTicketTime.After(time.Now()) {
		remain.remark = t.saleTicketTime.Format(constYMdHmFormat) + " 开售"
		return
	}
	seatBit := countSeatBit(depIdx, arrIdx)
	var noSeatCount uint8
	for i:=0; i<len(t.carTypesIdx) - 1; i++ {
		start, end := t.carTypesIdx[i], t.carTypesIdx[i + 1]
		if start == end {
			continue
		}
		availableSeatCount := 0
		for j:=start; j<end && availableSeatCount < constMaxAvaliableSeatCount; j++ {
			for k:=0; k<len(t.cars[j].seats) && availableSeatCount < constMaxAvaliableSeatCount; k++ {
				if t.cars[j].seats[0].IsAvailable(seatBit, isAdult) {
					availableSeatCount++
				}
			}
			if noSeatCount < constMaxAvaliableSeatCount {
				noSeatCount += t.cars[j].getAvailableNoSeatCount(seatBit, depIdx, arrIdx)
			}
		}		
		remain.availableSeatCount[i] = availableSeatCount
	}
	remain.availableSeatCount[len(remain.availableSeatCount) - 1] = int(noSeatCount)
	return
}

//////////////////////////////////////////////
///          车厢结构体及其对应方法          ///
//////////////////////////////////////////////
type car struct{
	// 车厢编号
	carNum uint8
	// 车厢的座位类型
	seatType string
	// 车厢的所有座位
	seats []*seat
	// 车厢内站票，无站票时此长度为零
	noSeats []*seat
	// 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	eachRouteTravelerCount []uint8
	// 读写锁，用于保护各路段乘客人数字段
	sync.RWMutex
}

// getAvailableSeat 获取可预订的座位,是否获取成功标记,是否为拼凑的站票标记
func (c *car)getAvailableSeat(seatBit uint64, isAdult, acceptNoSeat bool, depIdx, arrIdx uint8)(s *seat, ok, isMedley bool) {
	for i:=0; i<len(c.seats); i++ {
		if c.seats[i].IsAvailable(seatBit, isAdult){
			return c.seats[i], true, false
		}
	}
	if !acceptNoSeat || len(c.noSeats) == 0 {
		return nil, false, false
	}
	for i:=0; i<len(c.noSeats); i++ {
		if c.noSeats[i].IsAvailable(seatBit, true){
			return c.noSeats[i], true, false
		}
	}
	// 下面开始查找拼凑的站票
	// 非站票数与站票数之和
	totalSeatCount := len(c.seats) + len(c.noSeats)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	defer c.RUnlock()
	for i:=depIdx; i<arrIdx; i++ {
		if c.eachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.eachRouteTravelerCount[i]
		}
	}
	if totalSeatCount - int(maxTravelerCountInRoute) > 0 {
		s = &seat{seatNum:"", isAdult: true, seatBit: seatBit}
		return s, true, true
	}
	return nil, false, false
}

func (c *car)getAvailableNoSeatCount(seatBit uint64, depIdx, arrIdx uint8) uint8 {
	// 车厢未设置站票时，直接返回 0
	if len(c.noSeats) == 0 {
		return 0
	}
	// 非站票数与站票数之和
	totalSeatCount := len(c.seats) + len(c.noSeats)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	defer c.RUnlock()
	for i:=depIdx; i<arrIdx; i++ {
		if c.eachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.eachRouteTravelerCount[i]
		}
	}
	return uint8(totalSeatCount) - maxTravelerCountInRoute
}

// 某座位被占用
func (c *car)occupySeat(depIdx, arrIdx uint8)bool{
	noSeatLen := len(c.noSeats)
	if noSeatLen == 0{
		return true
	}
	c.Lock()
	defer c.Unlock()
	seatLen := len(c.seats)
	var maxCount uint8
	for i:=depIdx; i<arrIdx; i++ {
		if c.eachRouteTravelerCount[i] > maxCount {
			maxCount = c.eachRouteTravelerCount[i]
		}
	}
	if maxCount >= uint8(seatLen + noSeatLen) {
		return false
	}
	for i:=depIdx; i<arrIdx; i++ {
		c.eachRouteTravelerCount[i]++
	}
	return true
}

// 某座位被释放
func (c *car)releaseSeat(depIdx, arrIdx uint8){
	if len(c.noSeats) != 0 {
		 c.Lock()
		defer c.Unlock()
		for i:=depIdx; i<arrIdx; i++ {
			c.eachRouteTravelerCount[i]--
		}
	}
}

//////////////////////////////////////////////
///          座位结构体及其对应方法          ///
//////////////////////////////////////////////
type seat struct{
	// 座位号
	seatNum string
	// 是否成人票
	isAdult bool
	// 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	// 为什么用64位？ 因为途经站点最多的车次有57站。。。
	seatBit uint64
	// 锁，订票与退票均需要锁
	sync.Mutex
}

// IsAvailable 根据路段和乘客类型判断能否订票
func (s *seat)IsAvailable(seatBit uint64, isAdult bool) bool{
	return (s.isAdult == isAdult) && (s.seatBit ^ seatBit == s.seatBit + seatBit)
}

// Book 订票
func (s *seat)Book(seatBit, tranFullSeatBit uint64, isAdult bool)bool{
	s.Lock()
	defer s.Unlock()
	if !s.IsAvailable(seatBit, isAdult) {
		return false
	}
	s.seatBit ^= seatBit
	return true
}

// Refund 退票或取消订单，释放座位对应路段的资源
func (s *seat)Release(seatBit uint64){
	s.Lock()
	defer s.Unlock()
	s.seatBit ^= (^seatBit)
}

func countSeatBit(depIdx, arrIdx uint8) (result uint64) {
	for i:=depIdx; i<= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}