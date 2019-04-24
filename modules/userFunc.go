package modules

import (
	"time"
)

// QueryResidualTicketInfo 获取所选出发日期中， 经过出发站、目的站的车次及其各类余票数量
func QueryResidualTicketInfo(depStationName, arrStationName, depDate string, isStudent bool) (result []string) {
	depS, arrS := getStationInfoByName(depStationName), getStationInfoByName(arrStationName)
	if depS == nil || arrS == nil {
		return
	}
	matchTrans := getViaTrans(depS, arrS)
	queryDate, _ := time.Parse(ConstYmdFormat, depDate)
	resultCh, count := make(chan *ResidualTicketInfo, len(matchTrans)), 0
	for i := 0; i < len(matchTrans); i++ {
		depIdx, arrIdx, tdate, ok := matchTrans[i].IsMatchQuery(depS, arrS, queryDate)
		if !ok {
			continue
		}
		count++
		go func(t *TranInfo, depIdx, arrIdx uint8, date string) {
			rti := buildResidualTicketInfo(t, depIdx, arrIdx, date, isStudent)
			resultCh <- rti
		}(matchTrans[i], depIdx, arrIdx, tdate)
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
	tran, exist := getTranInfo(tranNum, date)
	if !exist {
		return
	}
	result = make([]TimetableResult, 0, len(tran.Timetable))
	for i, v := range tran.Timetable {
		r := TimetableResult{Name: v.StationName, DepTime: v.getStrDepTime(), ArrTime: v.getStrArrTime(), StayTime: v.getStrStayTime()}
		if i == 0 {
			r.ArrTime = ConstStrNullTime
			r.StayTime = ConstStrNullTime
		}
		if i == len(tran.Timetable)-1 {
			r.DepTime = ConstStrNullTime
			r.StayTime = ConstStrNullTime
		}
		result = append(result, r)
	}
	return
}

// QuerySeatPrice 查询票价
func QuerySeatPrice(tranNum string, date time.Time, depIdx, arrIdx uint8) (result map[string]float32) {
	if tran, exist := getTranInfo(tranNum, date); exist {
		result = tran.getSeatPrice(depIdx, arrIdx)
	}
	return
}

func countSeatBit(depIdx, arrIdx uint8) (result int64) {
	for i := depIdx; i <= arrIdx; i++ {
		result ^= 1 << i
	}
	return
}
