package modules

import (
	"sort"
	"time"
)

// QueryResidualTicketInfo 获取所选出发日期中， 经过出发站、目的站的车次及其各类余票数量
func QueryResidualTicketInfo(depStationName, arrStationName string, queryDate time.Time, isStudent bool) (result []string) {
	depS := getStationInfoByName(depStationName)
	depTrans, exist := cityTranMap[depS.CityCode]
	if !exist {
		return
	}
	arrS := getStationInfoByName(arrStationName)
	resultCh := make(chan *ResidualTicketInfo, 20)
	defer close(resultCh)
	matchTranCount := 0
	for i := 0; i < len(depTrans); i++ {
		depIdx, arrIdx, depDate, ok := depTrans[i].IsMatchQuery(depS, arrS, queryDate)
		if !ok {
			continue
		}
		matchTranCount++
		go func(ti *TranInfo, depIdx, arrIdx uint8, date string) {
			rti := newResidualTicketInfo(ti, depIdx, arrIdx)
			tran := getScheduleTran(ti.TranNum, date)
			rti.setScheduleInfo(tran, isStudent)
			resultCh <- rti
		}(depTrans[i], depIdx, arrIdx, depDate)
	}
	resultList := make([]*ResidualTicketInfo, matchTranCount)
	for i := 0; i < matchTranCount; i++ {
		resultList[i] = <-resultCh
	}
	// 按发车时间排序
	sort.Sort(residualTicketSort(resultList))
	result = make([]string, matchTranCount)
	for i := 0; i < len(resultList); i++ {
		result[i] = resultList[i].toString()
	}
	return
}

// TimetableResult 查询时刻表的结果
type TimetableResult struct {
	Name     string // 车站名
	DepTime  string // 出发时间
	ArrTime  string // 到达时间
	StayTime string // 停留时间
}

// QueryTimetable 查询时刻表
func QueryTimetable(tranNum string, date time.Time) (result []TimetableResult) {
	tran := getTranInfo(tranNum, date)
	result = make([]TimetableResult, len(tran.Timetable))
	for _, v := range tran.Timetable {
		r := TimetableResult{Name: v.StationName, DepTime: v.getStrDepTime(), ArrTime: v.getStrArrTime(), StayTime: v.getStrStayTime()}
		result = append(result, r)
	}
	return
}

// QuerySeatPrice 查询票价
func QuerySeatPrice(tranNum string, date time.Time, depIdx, arrIdx uint8) (result map[string]float32) {
	tran := getTranInfo(tranNum, date)
	result = tran.getSeatPrice(depIdx, arrIdx)
	return
}

func countSeatBit(depIdx, arrIdx uint8) (result uint64) {
	for i := depIdx; i <= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}
