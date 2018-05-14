package modules

import (
	"database/sql"
	"fmt"
)

// 车站集合
var stations []station

func initStation(db *sql.DB) {
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
		rows.Scan(&s.StationName, &s.StationCode, &s.CityCode)
		stations = append(stations, *s)
	}
}

type station struct {
	StationName string
	StationCode string
	CityCode    string
}

// 根据站点名，找出站点编码与城市编码
func getStationInfoByName(stationName string) *station {
	for i := 0; i < len(stations); i++ {
		if stations[i].StationName == stationName {
			return &stations[i]
		}
	}
	return nil
}
