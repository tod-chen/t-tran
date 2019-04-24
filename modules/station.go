package modules

import (
	"fmt"
	"sort"
	"strings"
)

// 车站集合
var stations stationCfgs

type stationCfgs []Station

func (sc stationCfgs) Len() int {
	return len(sc)
}

func (sc stationCfgs) Less(i, j int) bool {
	return -1 == strings.Compare(sc[i].StationName, sc[j].StationName)
}

func (sc stationCfgs) Swap(i, j int) {
	sc[i], sc[j] = sc[j], sc[i]
}

func initStation() {
	db.Where("is_passenger = 1").Find(&stations)
	sort.Sort(stations)
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
	idx := sort.Search(len(stations), func(i int) bool {
		return -1 != strings.Compare(stations[i].StationName, stationName)
	})
	if idx != -1 {
		return &stations[idx]
	}
	return nil
}
