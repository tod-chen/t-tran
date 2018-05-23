package modules

import (
	"strconv"
	"time"
)

// Route 时刻表信息
type Route struct {
	StationName     string    `bson:"stationName"`     //车站名
	StationCode     string    `bson:"stationCode"`     //车站编码
	CityCode        string    `bson:"cityCode"`        //城市编码
	DepTime         time.Time `bson:"depTime"`         //出发时间
	ArrTime         time.Time `bson:"arrTime"`         //到达时间
	CheckTicketGate string    `bson:"checkTicketGate"` //检票口
	Platform        int       `bson:"platform"`        // 乘车站台
}

const (
	constStrNullTime = "----"
)

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
