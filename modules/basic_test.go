package modules

import (
	"sort"
	"testing"
	"time"
)

func TestStructTranInfo(t *testing.T) {
	initTranInfo()
	if len(tranInfos) == 10425 {
		t.Log("trans info pass")
	} else {
		t.Error("trans info fail")
	}

	if sort.IsSorted(tranCfgs(tranInfos)) {
		t.Log("trans sorted pass")
	} else {
		t.Error("trans sorted fail")
	}

	if len(cityTranMap) == 620 {
		t.Log("city tran map pass")
	} else {
		t.Error("city tran map fail")
	}

	if len(cityTranMap["wuhan"]) == 1263 {
		t.Log("city tran map key pass")
	} else {
		t.Error("city tran map key fail")
	}

	tran, ok := getTranInfo("C2222", time.Now())
	if ok && len(tran.Timetable) == 3 &&
		tran.Timetable[0].StationName == "天津" &&
		tran.Timetable[2].StationName == "北京南" {
		t.Log("getTranInfo pass")
	} else {
		t.Error("getTranInfo fail")
	}

	if tran.isIntercity() {
		t.Log("isIntercity pass")
	} else {
		t.Error("isIntercity fail")
	}

	tran.getFullInfo()
	if len(tran.SeatPriceMap) == 0 {
		t.Log("getFullInfo pass")
	} else {
		t.Error("getFullInfo fail")
	}

	if ok, msg := tran.Save(); ok && msg == "" {
		t.Log("Save & initTimetable pass")
	} else {
		t.Error("Save & initTimetable fail")
	}

	seatPrice := tran.getSeatPrice(0, 1)
	if len(seatPrice) == 0 {
		t.Log("getSeatPrice pass")
	} else {
		t.Error("getSeatPrice fail")
	}

	depS := &Station{
		StationCode: "TJP",
		CityCode:"tianjin"}
	arrS := &Station {
		StationCode: "WWP",
		CityCode:"tianjin"}
	date := time.Now()
	if _, _, _, ok := tran.IsMatchQuery(depS, arrS, date); ok {
		t.Log("IsMatchQuery pass")
	} else {
		t.Error("IsMatchQuery fail")
	}

	depS.StationCode = "VNP"
	depS.CityCode = "beijing"
	if _, _, _, ok := tran.IsMatchQuery(depS, arrS, date); !ok {
		t.Log("IsMatchQuery pass")
	} else {
		t.Error("IsMatchQuery fail")
	}
}

func TestStructRoute(t *testing.T) {
	dt, _ := time.Parse(ConstYMdHmsFormat, "2018-01-01 13:08:00")
	at, _ := time.Parse(ConstYMdHmsFormat, "2018-01-01 13:03:00")
	r := &Route {
		DepTime: dt,
		ArrTime: at,
	}
	if depTime := r.getStrDepTime(); depTime == "13:08" {
		t.Log("getStrDepTime pass")
	} else {
		t.Error("getStrDepTime fail")
	}
	
	if arrTime := r.getStrArrTime(); arrTime == "13:03" {
		t.Log("getStrArrTime pass")
	} else {
		t.Error("getStrArrTime fail")
	}

	if stayTime := r.getStrStayTime(); stayTime == "5" {
		t.Log("getStrStayTime pass")
	} else {
		t.Error("getStrStayTime fail")
	}
}

func TestStructCar(t *testing.T) {
	c := &Car{
		TranType:"G",
		SeatType:"Test",
		SeatCount: 2,
		NoSeatCount:10,
		Remark:"Car Test Unit",
		Seats: []Seat{
				Seat{
					CarID: 0,
					SeatNum:"1",
					IsStudent: true},
				Seat{
					CarID:0,
					SeatNum:"2",
					IsStudent:false},
			},
	}
	if ok, msg := c.Save(); ok && msg == "" {
		t.Log("Save pass")
	} else {
		t.Error("Save fail")
	}
}