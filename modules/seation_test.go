package modules

import "testing"

// TestInitStation 测试站点初始化方法
func TestInitStation(t *testing.T){
	initStation()
	if len(stations) == 2323 {
		t.Log("pass")
	} else {
		t.Error("fail")
	}
}

// TestGetStationInfoByName 测试根据站点名查找站点
func TestGetStationInfoByName(t *testing.T){
	if len(stations) == 0 {
		initStation()
	}
	s := getStationInfoByName("武汉")
	if s.StationCode == "WHN" && s.CityCode == "wuhan" {
		t.Log("pass")
	} else {
		t.Error("fail")
	}
}
