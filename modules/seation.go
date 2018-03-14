package modules

// 站点名称与城市代码关系
var stationMap map[string]string
var stations []station

func init(){
	// TODO: 初始化stations & stationMap
}

type station struct{
	stationName string
	stationCode string
	cityCode string
}

// GetCityCodeByStationName 根据站点名，找出城市编码
func GetCityCodeByStationName(stationName string)string{
	for name, cityCode := range stationMap {
		if name == stationName {
			return cityCode
		}
	}
	return ""
}

// GetRelationStations 根据城市编码，找出同市内的其他站点名及站点编码
func GetRelationStations(cityCode string)map[string]string{
	result := make(map[string]string)
	for _, item := range stations {
		if cityCode == item.cityCode {
			result[item.stationName] = item.stationCode
		}
	}
	return result
}