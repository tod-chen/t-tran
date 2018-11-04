package modules

import (
	"time"
)

// QueryResidualTicketInfo 获取所选出发日期中， 经过出发站、目的站的车次及其各类余票数量
func QueryResidualTicketInfo(depStationName, arrStationName string, queryDate time.Time, isStudent bool) (result []string) {
	depS, arrS := getStationInfoByName(depStationName), getStationInfoByName(arrStationName)
	matchTrans := getViaTrans(depS, arrS)
	resultCh, count := make(chan *ResidualTicketInfo, 20), 0
	for i := 0; i < len(matchTrans); i++ {
		depIdx, arrIdx, depDate, ok := matchTrans[i].IsMatchQuery(depS, arrS, queryDate)
		if !ok {
			continue
		}
		count++
		go func(t *TranInfo, depIdx, arrIdx uint8, date string) {
			rti := newResidualTicketInfo(t, depIdx, arrIdx)
			tran := getScheduleTran(t.TranNum, date)
			rti.setScheduleInfo(tran, isStudent)
			resultCh <- rti
		}(matchTrans[i], depIdx, arrIdx, depDate)
	}
	result = make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = (<-resultCh).toString()
	}
	close(resultCh)
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
