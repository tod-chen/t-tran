package modules

import (
	"strconv"
	"time"
)

type route struct{
	stationName string
	checkTicketGate string
	stationCode string
	cityCode string
	depTime time.Time
	arrTime time.Time
}

const (
	constStrNullTime = "----"
)

func (r *route)getStrDep()string{
	if r.depTime.Year() == 1 {
		return constStrNullTime
	}
	return r.depTime.Format(constHmFormat)
}

func (r *route)getStrArr()string{	
	if r.arrTime.Year() == 1 {
		return constStrNullTime
	}
	return r.arrTime.Format(constHmFormat)
}

func (r *route)getStrStayTime()string{
	if r.depTime.Year() == 1 || r.arrTime.Year() == 1 {
		return constStrNullTime
	}
	return strconv.FormatFloat(r.depTime.Sub(r.arrTime).Minutes(), 'e', 0, 64) 
}