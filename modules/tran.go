package modules

import (
	"strings"
	"strconv"
	"time"
)

const (
	// 查询时距离当前时间多长时间内发车的车次不予显示 单位：分钟
	queryTranDelay = 20
	// 未完成订单有效时间 单位：分钟
	unpayOrderAvaliableTime = 45
	// 查询余票数量的最大值 超过此值时，显示“有”
	maxAvaliableSeatCount = 100
	// 站票最大数量
	maxNoSeatCount = 10000
	// 小时:分钟 格式
	hmFormat = "15:04"
	ymdFormat = "2006-01-02"

	// SeatTypeNoSeat 站票的座位类型
	SeatTypeNoSeat = "NS"
)

// AllTrans 所有车次
var AllTrans []tran

func init(){
	// TODO: 初始化 AllTrans
}

type tran struct{
	id int
	tranNum string
	tranType string
	departureDate time.Time
	saleTicketTime time.Time
	notSaleRemark string
	routeTimetable []route
	// seatTypesIndex  座位集合中各种座次的起始索引
	// 依次是：商务座、一等座、二等座、高级软卧、软卧、动卧、硬卧、软座、硬座、无座
	seatTypesIndex []uint16
	seats []*seat
	noSeatEachRouteTravelerCount []uint16
}

// IsMatchStationAndTime 判断当前车次在站点和时间上是否匹配
func (t *tran)IsMatchStationAndTime(depCityCode, arrCityCode string, depDate time.Time) (depIndex, arrIndex uint, ok bool){
	depI := -1

	year, month, day := depDate.Date()
	now := time.Now().Add(queryTranDelay * time.Minute)

	// TODO: 当某车次的路线经过某城市的两个站，该怎么匹配？ 当前算法是匹配第一个，与12306逻辑一致 12306这里算是一个bug
	for i, item := range t.routeTimetable{
		if depI == -1 && item.cityCode == depCityCode {
			y, m, d := item.depTime.Date()
			// 日期不匹配， 放在这里是因为有些车次是跨日的，有些车次可能发车的次日才能到达某个站点
			if d != day || m != month || y != year {
				return
			}
			// 早于当前时间的车次，默认已经发车，赶不上了
			if item.depTime.Before(now) {
				return
			}
			depI = i
			depIndex = uint(i)
		}
		if depI != -1 && item.cityCode == arrCityCode {
			arrIndex = uint(i)
			ok = true
			return
		}
	}
	return
}

// GetAvailableSeatCount 获取各座次余票数
func (t *tran)GetAvailableSeatCount(depIndex, arrIndex uint, isAdult bool)string{
	first, dep, arr, last := t.routeTimetable[0], t.routeTimetable[depIndex], t.routeTimetable[arrIndex], t.routeTimetable[len(t.routeTimetable)-1]
	dur := arr.arrTime.Sub(dep.depTime)
	list := []string{ t.tranNum,  // 车次
		first.depTime.Format(ymdFormat), first.stationCode, // 起点站发车日期， 起点站编码
		dep.depTime.Format(hmFormat), dep.stationCode, // 出发站 发车 时:分， 出发站编码
		arr.arrTime.Format(hmFormat), arr.stationCode,	// 目的站 到站 时:分， 目的站编码
		strings.Replace(strings.Split(dur.String(), "m")[0], "h", ":", 1), last.stationCode } // 耗时 时:分， 终点站编码
	array := make([]string, 10)
	if t.notSaleRemark != "" {
		return strings.Join(list, "|") + "|" + strings.Join(array, "|") + "|0|" + t.notSaleRemark
	}
	n := len(t.seatTypesIndex)
	seatMatch := getSeatMatch(depIndex, arrIndex)
	var count int
	canBookFlag := "0"
	for i:=0; i < n - 1; i++{
		start, end := t.seatTypesIndex[i], t.seatTypesIndex[i + 1]
		if start == end {
			continue
		}
		if i != n - 2 { // 非站票
			count = int(t.getSeatAvaliableCount(seatMatch, start, end, isAdult))
		} else{ // 站票具有流动性
			count = int(t.getNoSeatAvaliableCount(seatMatch, depIndex, arrIndex))
		}
		if canBookFlag == "0" && count != 0 {
			canBookFlag = "1"
		}
		array[i] = strconv.Itoa(count)
	}
	return strings.Join(list, "|") + "|" + strings.Join(array, "|") + "|" + canBookFlag + "|"
}

func (t *tran)getSeatAvaliableCount(seatMatch uint32, start, end uint16, isAdult bool) (count int16) {
	// 没有某座次的座位
	if start == end {
		return -1
	}
	for i:=start; i<end; i++{
		if count < maxAvaliableSeatCount && t.seats[i].IsAvailable(seatMatch, isAdult) {
			count++
		}
	}
	return count
}

func (t *tran)getNoSeatAvaliableCount(seatMatch uint32, depIndex, arrIndex uint) (count int16) {
	typeIndex := uint(len(t.seatTypesIndex) - 2)
	var seatableTypeIndex uint
	for i:=typeIndex; i>0; i-- {
		if t.seatTypesIndex[i] != t.seatTypesIndex[i-1] {
			seatableTypeIndex = i - 1
			break
		}
	}
	noSeatMap := map[uint16]uint16 {
		t.seatTypesIndex[seatableTypeIndex] : t.seatTypesIndex[seatableTypeIndex + 1],
		t.seatTypesIndex[typeIndex] : t.seatTypesIndex[typeIndex + 1] }
	eachRouteAvailableCount := make([]int16, arrIndex - depIndex)
	for i:=depIndex; i<arrIndex; i++ {		
		eachRouteAvailableCount[i-depIndex] -= int16(t.noSeatEachRouteTravelerCount[depIndex + i])
	}
	for start, end := range noSeatMap {
		for i:= start; i<end; i++ {
			if t.seats[i].isFull {
				continue
			}
			for j:=depIndex; j<arrIndex; j++ {
				if t.seats[i].seatRoute ^ (1 << j) == 0 {
					eachRouteAvailableCount[j-depIndex]++
				}
			}
		}
	}
	count = maxNoSeatCount
	for _, val := range eachRouteAvailableCount {
		if val < count {
			count = val
		}
	}
	return
}

func (t *tran)getOneAvailableNoSeat(seatMatch uint32, depIndex, arrIndex uint)*seat{
	typeIndex := len(t.seatTypesIndex) - 2
	start, end := t.seatTypesIndex[typeIndex], t.seatTypesIndex[typeIndex + 1]
	for i:=start; i<end; i++ {
		if t.seats[i].IsAvailable(seatMatch, false) {
			return t.seats[i]
		}
	}
	count := t.getNoSeatAvaliableCount(seatMatch, depIndex, arrIndex)
	if count > 0 {
		for i:=start; i<end; i++ {
			if t.seats[i].seatRoute ^ (1 << depIndex) == 0 {
				s := &seat{
					tranID: t.seats[i].tranID,
					carNum: t.seats[i].carNum,
					seatNum : "",
					seatType : SeatTypeNoSeat,
					isPutTogetherNoSeat: true }
				return s
			}
		}
	}
	return nil
}

func (t *tran)getDepAndArrIndexByStationName(depName, arrName string)(depIndex, arrIndex uint){
	tempDep := -1
	for i, r := range t.routeTimetable {
		if tempDep == -1 && r.stationName == depName {
			tempDep = i
			depIndex = uint(i)
		}
		if tempDep != -1 && r.stationName == arrName {
			arrIndex = uint(i)
			return
		}
	}
	return
}
