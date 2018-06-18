package modules

import (
	// "sort"
	"strconv"
	"strings"
	"time"
)

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

func newRemainingTickets(t *TranInfo, depIdx, arrIdx uint8) *RemainingTickets {
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
		//remark:             t.NotSaleRemark,
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
// func QueryMatchTransInfo(depStationName, arrStationName string, queryDate time.Time, isStudent bool) (result []string) {
// 	depS := getStationInfoByName(depStationName)
// 	depTrans, exist := cityTranMap[depS.CityCode]
// 	if !exist {
// 		return
// 	}
// 	arrS := getStationInfoByName(arrStationName)
// 	availableTime := time.Now().Add(constQueryTranDelay * time.Minute)
// 	resultCh := make(chan *RemainingTickets, 20)
// 	defer close(resultCh)
// 	matchTranCount := 0
// 	for i := 0; i < len(depTrans); i++ {
// 		depIdx, arrIdx, ok := depTrans[i].IsMatchStation(depS, arrS)
// 		if !ok {
// 			continue
// 		}
// 		// 根据起点站和出发站之间的时间跨度，以及查询日期，算出该列车的发车日期
// 		firstStationDepDate := depTrans[i].RouteTimetable[0].DepTime
// 		depStationDate := depTrans[i].RouteTimetable[depIdx].DepTime
// 		subDays := int(depStationDate.Sub(firstStationDepDate).Hours() / 24)
// 		dateKey := queryDate.AddDate(0, 0, -subDays).Format(constYmdFormat)
// 		// 遍历当前列车出发日的所有列车，找出匹配项，这里需要做优化，改用二分法查找(需要对tranList排序)
// 		trans := tranScheduleMap[dateKey]
// 		for i := 0; i < len(trans); i++ {
// 			if depTrans[i].TranNum != trans[i].TranNum {
// 				continue
// 			}
// 			depTime := trans[i].RouteTimetable[depIdx].DepTime
// 			// 发车时间在所查询的日期内, 跳出外层循环
// 			if queryDate.Before(depTime) && depTime.Before(queryDate.Add(constOneDayDuration)) {
// 				ok = false
// 				// 发车前20分钟内，不予查询
// 				if !depTime.Before(availableTime) {
// 					break
// 				}
// 				matchTranCount++
// 				goPool.Take()
// 				go func(tranInfo *TranInfo, depIdx, arrIdx uint8) {
// 					defer goPool.Return()
// 					resultCh <- trans[i].GetTranInfoAndSeatCount(depIdx, arrIdx, tranInfo.carTypesIdx, isStudent)
// 				}(depTrans[i], depIdx, arrIdx)
// 				break
// 			}
// 		}
// 	}
// 	resultList := make([]*RemainingTickets, matchTranCount)
// 	for i := 0; i < matchTranCount; i++ {
// 		resultList[i] = <-resultCh
// 	}
// 	sort.Sort(remainingTicketSort(resultList))
// 	result = make([]string, matchTranCount)
// 	for i := 0; i < len(resultList); i++ {
// 		result[i] = resultList[i].toString()
// 	}
// 	return
// }

// QueryRouteResult 查询时刻表的结果
type QueryRouteResult struct {
	name     string // 车站名
	depTime  string // 出发时间
	arrTime  string // 到达时间
	stayTime string // 停留时间
}

// QueryRoutetable 查询时刻表
// func QueryRoutetable(tranNum string) (result []QueryRouteResult) {
// 	for i := 0; i < len(tranList); i++ {
// 		if tranList[i].TranNum == tranNum {
// 			routeLen := len(tranList[i].RouteTimetable)
// 			result = make([]QueryRouteResult, routeLen)
// 			for j := 0; j < routeLen; j++ {
// 				routeInfo := QueryRouteResult{
// 					name:     tranList[i].RouteTimetable[j].StationName,
// 					depTime:  tranList[i].RouteTimetable[j].getStrDep(),
// 					arrTime:  tranList[i].RouteTimetable[j].getStrArr(),
// 					stayTime: tranList[i].RouteTimetable[j].getStrStayTime(),
// 				}
// 				result = append(result, routeInfo)
// 			}
// 			break
// 		}
// 	}
// 	return
// }

// QuerySeatPrice 查询票价
func QuerySeatPrice(tranNum string, depIdx, arrIdx uint8) (result map[string]float32) {
	// for i := 0; ; i++ {
	// 	if tranList[i].TranNum == tranNum {
	// 		result = tranList[i].getSeatPrice(depIdx, arrIdx)
	// 		break
	// 	}
	// }
	return
}


func countSeatBit(depIdx, arrIdx uint8) (result uint64) {
	for i := depIdx; i <= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}
