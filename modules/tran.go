package modules

import (
	"sync"
	"strings"
	"strconv"
	"time"
	"fmt"
	"database/sql"
)

const (
	// 时间格式 小时:分钟 
	constHmFormat = "15:04"
	// 时间格式 年-月-日
	constYmdFormat = "2006-01-02"
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
	// 所有接受订票的列车集合
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
			fmt.Println(err)
			break
		}
		t := tran{}
		t.tranNum = tranNum
		t.departureTime, err = time.Parse(constHmFormat, depTime) 
		t.runDays = uint8(runDays)
		tranList = append(tranList, t)
	}
	for i:=0; i<len(tranList); i++ {
		// 初始化所有列车的组合字段
		routes := make([]route, 0, 10)	// 根据车次号从路线表中读取
		queryRoute := "select stationName, stationCode, cityCode, arrTime, depTime from routes where tranNum = '" + tranList[i].tranNum + "' order by stationIndex"
		routeRows, _ := db.Query(queryRoute)
		defer routeRows.Close()
		for routeRows.Next() {
			r := new(route)
			routeRows.Scan(&r.stationName, &r.stationCode, &r.cityCode, &r.arrTime, &r.depTime)
			routes = append(routes, *r)
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

func initTranScheduleMap(){	
	fmt.Println("begin init tranScheduleMap")
	defer fmt.Println("end init tranScheduleMap")
	tranScheduleMap = make(map[string]([]tran), constDays)
	now := time.Now()
	for i:=0; i<len(tranList); i++ {
		for d:=0; d<constDays; d++ {
			date := now.AddDate(0, 0, d).Format(constYmdFormat)
			trans, exist := tranScheduleMap[date]
			if !exist {
				trans = make([]tran, 0, constTranCount)
			}
			tranScheduleMap[date] = append(trans, tranList[i])			
		}
	}
}
//////////////////////////////////////////////
///          列车结构体及其方法              ///
//////////////////////////////////////////////


// QueryMatchTransInfo 获取所选出发日期中， 经过出发站、目的站的车次及其各类余票数量
func QueryMatchTransInfo(depCityCode, arrCityCode string, depDate time.Time, isAdult bool)(result []string) {
	
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
	for i:=0; ; i++ {
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

type tran struct{
	// 车次ID
	id int
	// 车次号
	tranNum string
	// 发车时间
	departureTime time.Time
	// 售票时间
	saleTicketTime time.Time
	// 不售票的说明	路线调整啥的
	notSaleRemark string
	// 时刻表
	routeTimetable []route
	// 各类席位在各路段的价格
	seatPriceMap map[string]([]float32)
	// 车厢
	cars []car
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
func (t *tran)IsMatchStation(depCityCode, arrCityCode string) (depIdx, arrIdx uint8, ok bool){
	depI := -1
	now := time.Now().Add(constQueryTranDelay * time.Minute)

	// TODO: 当某车次的路线经过某城市的两个站，该怎么匹配？ 当前算法是匹配第一个，与12306逻辑一致 12306这里算是一个bug
	routeLen := len(t.routeTimetable)
	for i:=0; i<routeLen; i++ {
		if depI == -1 && t.routeTimetable[i].cityCode == depCityCode {
			// constQueryTranDelay 分钟内将发车的车次不予显示，默认赶不上车了
			if t.routeTimetable[i].depTime.Before(now) {
				return
			}
			depI = i
			depIdx = uint8(i)
		}
		if depI != -1 && t.routeTimetable[i].cityCode == arrCityCode {
			arrIdx = uint8(i)
			ok = true
			return
		}
	}
	return
}

// GetTranInfoAndSeatCount 返回车次信息、起止站在时刻表中的索引、各类座位余票数、可售标记、不售票说明
func (t *tran)GetTranInfoAndSeatCount(depIdx, arrIdx uint8, isAdult bool)string{
	strDepIdx, strArrIdx := strconv.Itoa(int(depIdx)), strconv.Itoa(int(arrIdx))
	first, dep, arr, last := t.routeTimetable[0], t.routeTimetable[depIdx], t.routeTimetable[arrIdx], t.routeTimetable[len(t.routeTimetable)-1]
	dur := arr.arrTime.Sub(dep.depTime)
	list := []string{ t.tranNum,  // 车次
		first.depTime.Format(constYmdFormat), first.stationCode, // 起点站发车日期， 起点站编码
		dep.depTime.Format(constHmFormat), dep.stationCode, // 出发站 发车 时:分， 出发站编码
		arr.arrTime.Format(constHmFormat), arr.stationCode,	// 目的站 到站 时:分， 目的站编码
		strings.Replace(strings.Split(dur.String(), "m")[0], "h", ":", 1), last.stationCode } // 耗时 时:分， 终点站编码
	eachSeatTypeCounts := make([]string, constSeatTypeCount)
	if t.notSaleRemark != "" && time.Now().Before(t.saleTicketTime) {
		return strings.Join(list, "|") + "|" + strDepIdx + "|" + strArrIdx + "|" + strings.Join(eachSeatTypeCounts, "|") + "|0|" + t.notSaleRemark
	}

	n := len(t.carTypesIdx)
	seatMatch := countSeatBit(depIdx, arrIdx)
	var count, noSeatCount int
	noSeatCount = -1
	canBookFlag := "0"
	for i:=0; i < n - 1; i++ {
		start, end := t.carTypesIdx[i], t.carTypesIdx[i + 1]
		if start == end {
			continue
		}
		count = t.getSeatAvaliableCount(seatMatch, start, end, depIdx, arrIdx, isAdult, &noSeatCount)
		if canBookFlag == "0" && count > 0 {
			canBookFlag = "1"
		}
		eachSeatTypeCounts[i] = strconv.Itoa(count)
	}
	if canBookFlag == "0" && noSeatCount > 0 {
		canBookFlag = "1"
	}
	var strNoSeat string
	if noSeatCount != -1 {
		strNoSeat = strconv.Itoa(noSeatCount)
	}
	eachSeatTypeCounts[n-1] = strNoSeat
	return strings.Join(list, "|") + "|" + strDepIdx + "|" + strArrIdx + "|" + strings.Join(eachSeatTypeCounts, "|") + "|" + canBookFlag + "|"
}

func (t *tran)getSeatAvaliableCount(seatMatch uint64, carStartIdx, carEndIdx, depIdx, arrIdx uint8, isAdult bool, noSeatCount *int) (count int) {
	for i:=carStartIdx; i<carEndIdx; i++{
		if count < constMaxAvaliableSeatCount {
			for _, s := range t.cars[i].seats {
				if count < constMaxAvaliableSeatCount && 
					!s.isFull && s.IsAvailable(seatMatch, isAdult) {
					count++
				}
			}
		}
		noSeatLen := len(t.cars[i].noSeats)
		if noSeatLen == 0 {
			continue
		}
		if *noSeatCount < constMaxAvaliableSeatCount {
			for _, s := range t.cars[i].noSeats {
				// 可直接使用的站票
				if *noSeatCount < constMaxAvaliableSeatCount &&
					!s.isFull && s.IsAvailable(seatMatch, true) {
					(*noSeatCount)++
				}
			}
		}
	}
	// 直接可用的站票数量不足时，统计可拼凑的站票数量
	if *noSeatCount < constMaxAvaliableSeatCount {
		for i:=carStartIdx; i<carEndIdx; i++{
			if *noSeatCount >= constMaxAvaliableSeatCount {
				break
			}
			noSeatLen := len(t.cars[i].noSeats)
			if noSeatLen == 0 {
				continue
			}
			// 将坐票和站票合并
			seatLen := len(t.cars[i].seats)
			seats := make([]*seat, seatLen + noSeatLen)
			copy(seats, t.cars[i].noSeats)
			copy(seats[noSeatLen:], t.cars[i].seats)			
			// 出发地与目的地之间各路段旅客的人数
			eachRouteTravelerCount := make([]uint8, arrIdx - depIdx)
			for _, s := range seats {
				for i:=depIdx; i<arrIdx; i++ {
					if s.seatBit & (1 << i) != 0 {
						eachRouteTravelerCount[i-depIdx]++
					}
				}
			}
			maxCount := uint8(seatLen + noSeatLen)
			for i:=0; i<int(arrIdx-depIdx); i++ {
				if eachRouteTravelerCount[i] < maxCount {
					maxCount = eachRouteTravelerCount[i]
				}
			}
			*noSeatCount += (seatLen + noSeatLen - int(maxCount))
		}
	}
	return
}

// getOneAvailableNoSeat 获取一个无座的位置
func (t *tran)getOneAvailableNoSeat(seatBit uint64, depIdx, arrIdx uint8)*seat{
	carLen := len(t.cars)
	for i:=0; i<carLen; i++ {
		// 没有无座席位时，跳过
		if len(t.cars[i].noSeats) == 0 {
			continue
		}
		for _, s := range t.cars[i].noSeats {
			if !s.isFull && s.IsAvailable(seatBit, true) {
				return s
			}
		}
	}
	// 当各车厢均没有可直接预定的站票，则获取拼凑的站票
	for i:=0; i<carLen; i++ {
		// 没有无座席位时，跳过
		if len(t.cars[i].noSeats) == 0 {
			continue
		}
		// 将坐票和站票合并
		seatLen := len(t.cars[i].seats)
		noSeatLen := len(t.cars[i].noSeats)
		seats := make([]*seat, seatLen + noSeatLen)
		copy(seats, t.cars[i].noSeats)
		copy(seats[noSeatLen:], t.cars[i].seats)
		
		// 出发地与目的地之间各路段旅客的人数
		eachRouteTravelerCount := make([]uint8, arrIdx - depIdx)
		for _, s := range seats {
			for i:=depIdx; i<arrIdx; i++ {
				if s.seatBit & (1 << i) != 0 {
					eachRouteTravelerCount[i-depIdx]++
				}
			}
		}
		maxCount := uint8(seatLen + noSeatLen)
		for i:=0; i<int(arrIdx-depIdx); i++ {
			if eachRouteTravelerCount[i] < maxCount {
				maxCount = eachRouteTravelerCount[i]
			}
		}
		if (uint8(seatLen + noSeatLen) - maxCount) > 0 {
			for _, s := range seats {
				if s.seatBit & (1 << depIdx) == 0 {
					return s
				}
			}
		}
	}
	return nil
}

// func (t *tran)getDepAndArrIndexByStationName(depName, arrName string)(depIndex, arrIndex uint){
// 	tempDep := -1
// 	for i, r := range t.routeTimetable {
// 		if tempDep == -1 && r.stationName == depName {
// 			tempDep = i
// 			depIndex = uint(i)
// 		}
// 		if tempDep != -1 && r.stationName == arrName {
// 			arrIndex = uint(i)
// 			return
// 		}
// 	}
// 	return
// }


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
	// 锁，用于保护各路段乘客人数字段
	lock sync.Mutex
}

// getAvailableSeat 获取可预订的座位
func (c *car)getAvailableSeat(seatBit uint64, isAdult, acceptNoSeat bool, depIdx, arrIdx uint8)(s *seat, ok, isMedley bool) {
	for _, item := range c.seats {
		if !item.isFull && item.IsAvailable(seatBit, isAdult){
			return item, true, false
		}
	}
	if !acceptNoSeat {
		return nil, false, false
	}
	for _, item := range c.noSeats {
		if !item.isFull && item.IsAvailable(seatBit, true) {
			return item, true, false
		}
	}
	totalSeatCount := len(c.seats) + len(c.noSeats)
	availableSeatCount, tempSeatCount := totalSeatCount, 0
	for i:=depIdx; i<arrIdx; i++ {
		tempSeatCount = totalSeatCount - int(c.eachRouteTravelerCount[i])
		if tempSeatCount < availableSeatCount {
			availableSeatCount = tempSeatCount
		}
	}
	if availableSeatCount > 0 {
		for _, item := range c.seats {
			if !item.isFull && (item.seatBit & (1 << depIdx) == 0 ) {
				s = &seat{
					seatNum: "",
					isFull: false,
					isAdult: true,
				}
				return s, true, true
			}
		}
	}
	return nil, false, false
}

// 某座位被占用
func (c *car)occupySeat(depIdx, arrIdx uint8)bool{
	noSeatLen := len(c.noSeats)
	if noSeatLen == 0{
		return true
	}
	c.lock.Lock()
	defer c.lock.Unlock()
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
		 c.lock.Lock()
		defer c.lock.Unlock()
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
	// 是否全程都有人坐
	isFull bool
	// 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	// 为什么用64位？ 因为途经站点最多的车次有57站。。。
	seatBit uint64
	// 锁，订票与退票均需要锁
	sync.Mutex
}

// IsAvailable 根据路段和乘客类型判断能否订票
func (s *seat)IsAvailable(seatBit uint64, isAdult bool) bool{
	if s.isAdult != isAdult || s.isFull{
		return false
	}
	return s.seatBit ^ seatBit == s.seatBit + seatBit
}

// Book 订票
func (s *seat)Book(seatBit, tranFullSeatBit uint64, isAdult bool)bool{
	s.Lock()
	defer s.Unlock()
	if !s.IsAvailable(seatBit, isAdult) {
		return false
	}
	s.seatBit ^= seatBit
	if s.seatBit == tranFullSeatBit{
		s.isFull = true
	}
	return true
}

// Refund 退票或取消订单，释放座位对应路段的资源
func (s *seat)Release(seatBit uint64){
	s.Lock()
	defer s.Unlock()
	s.seatBit ^= (^seatBit)
	if s.isFull {
		s.isFull = false
	}
}

func countSeatBit(depIdx, arrIdx uint8) (result uint64) {
	for i:=depIdx; i<= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}