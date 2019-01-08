package modules

import "fmt"

// 车站集合
var stations []Station

func initStation() {
	db.Where("is_passenger = 1").Find(&stations)
	fmt.Println("init stations complete")
}

// Station 车站信息
type Station struct {
	ID            uint
	StationName   string // 车站名
	StationCode   string // 车站编码
	StationPinyin string // 车站拼音
	CityCode      string // 城市编码
	CityName      string // 城市名
	IsPassenger   bool   // 是否为客运站
}

// 根据站点名，找出站点编码与城市编码
func getStationInfoByName(stationName string) *Station {
	for i := 0; i < len(stations); i++ {
		if stations[i].StationName == stationName {
			return &stations[i]
		}
	}
	return nil
}
