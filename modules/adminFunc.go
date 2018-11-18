package modules

// QueryTrans 查询列车信息
func QueryTrans(tranNum, tranType string, page, pageSize int) (trans []TranInfo, count int) {
	q := db.Table("tran_infos").Where("tran_num like ? and tran_num REGEXP ?", "%"+tranNum+"%", "^"+tranType).Count(&count)
	q.Order("tran_num").Offset((page - 1) * pageSize).Limit(pageSize).Find(&trans)
	for i := 0; i < len(trans); i++ {
		trans[i].getFullInfo()
	}
	return
}

// GetTranDetail 获取车次明细
func GetTranDetail(tranID int) (t TranInfo) {
	db.Where("id = ?", tranID).First(&t)
	t.getFullInfo()
	return
}

// QueryCars 查询车厢信息
func QueryCars(seatType, tranType string, page, pageSize int) (cars []Car, count int) {
	q := db.Table("cars").Where("seat_type like ? and tran_type like ?", "%"+seatType+"%", "%"+tranType+"%").Count(&count)
	q.Order("tran_type").Offset((page - 1) * pageSize).Limit(pageSize).Find(&cars)
	return
}

// GetCarDetail 获取车厢明显
func GetCarDetail(carID int) (c Car) {
	db.Where("id = ?", carID).First(&c)
	var seats []Seat
	db.Where("car_id = ?", carID).Order("seat_num").Find(&seats)
	c.Seats = seats
	return
}

// QuerySchedule 查询排班信息
func QuerySchedule(departureDate, tranNum string, page, pageSize int) (schedules []ScheduleTran, count int) {
	
	q := db.Table("schedule_trans").Where("departure_date = ? and tran_num like ?", departureDate, "%"+tranNum+"%").Count(&count)
	q.Order("tran_num").Offset((page - 1) * pageSize).Limit(pageSize).Find(&schedules)
	return
}

// GetScheduleDetail 获取排班明细
func GetScheduleDetail(scheduleID int) (schedule ScheduleTran) {
	db.Where("id = ?", scheduleID).First(&schedule)
	var cars []ScheduleCar
	db.Where("schedule_tran_id = ?", scheduleID).Order("car_num").Find(&cars)
	schedule.Cars = cars
	return
}

// QueryStations 查询车站信息
func QueryStations(stationName, cityName string, page, pageSize int) (stations []Station, count int) {
	q := db.Table("stations").Where("station_name like ? and city_name like ?", "%"+stationName+"%", "%"+cityName+"%").Count(&count)
	q.Order("city_code").Offset((page - 1) * pageSize).Limit(pageSize).Find(&stations)
	return
}

// GetStationDetail 获取车站明细
func GetStationDetail(stationID int) (s Station) {
	db.Where("id = ?", stationID).First(s)
	return
}
