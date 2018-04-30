package modules

import (	
	"fmt"
	"database/sql"
)
// 车站集合
var stations []station

func initStation(db *sql.DB){
	fmt.Println("begin init stations")
	defer fmt.Println("end init stations")
	stations = make([]station, 0, 2400)
	query := "select stationName, stationCode, cityCode from stations where isPassenger is not null"
	rows, err := db.Query(query)
	if err != nil {
		panic("query error")
	}
	defer rows.Close()
	for rows.Next() {
		s := new(station)
		rows.Scan(&s.stationName, &s.stationCode, &s.cityCode)
		stations = append(stations, *s)
	}
}

type station struct{
	stationName string
	stationCode string
	cityCode string
}

// 根据站点名，找出站点编码与城市编码
func getCityCodeByStationCode(stationCode string) string {
	for _, s := range stations {
		if s.stationCode == stationCode {
			return s.cityCode
		}
	}
	return ""
}