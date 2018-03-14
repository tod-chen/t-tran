package modules

import "time"

type route struct{
	stationName string
	stationCode string
	cityCode string
	depTime time.Time
	arrTime time.Time
}