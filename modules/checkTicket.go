package modules

import "time"

// 已支付的全部有效订单
var allOrders []Order

// checkTicket 校验订单，以便释放无效订单所占用的资源 或 暴露冲突的订单
func checkTicket() {
	date := time.Now()
	for _, t := range tranInfos {
		for i := 0; i < constDays; i++ {
			dateStr := date.AddDate(0, 0, -i).Format(ConstYmdFormat)
			checkTranTicket(t.TranNum, dateStr)
		}
	}
}

// checkTranTicket 校验具体某趟车次的订单
func checkTranTicket(tranNum, date string) {
	st := getScheduleTran(tranNum, date)
	copySt := *st
	for ci := 0; ci < len(copySt.Cars); ci++ {
		if copySt.Cars[ci].NoSeatCount != 0 {
			for ei := 0; ei < len(copySt.Cars[ci].EachRouteTravelerCount); ei++ {
				copySt.Cars[ci].EachRouteTravelerCount[ei] = 0
			}
		}
		for si := 0; si < len(copySt.Cars[ci].Seats); si++ {
			copySt.Cars[ci].Seats[si].SeatBit = 0
		}
	}
	validTicketStatus := []uint8{constTicketPaid, constTicketIssued, constTicketChangePaid, constTicketChangeIssued}
	var tickets []Ticket
	db.Where("tran_num = ? and tran_dep_date = ? and status in (?)", tranNum, date, validTicketStatus).Find(&tickets)
	for _, t := range tickets {
		car, seat := getCarAndSeat(&copySt, t.CarNum, t.SeatType, t.SeatNum)
		seatBit := countSeatBit(t.DepStationIdx, t.ArrStationIdx)

		car.occupySeat(t.DepStationIdx, t.ArrStationIdx)
		if seat != nil {
			if ok := seat.Book(seatBit, t.IsStudent); !ok {
				notify := &notifyAdminInfo{
					date:       t.TranDepDate,
					tranNum:    t.TranNum,
					carNum:     t.CarNum,
					depStation: t.DepStation,
					arrStation: t.ArrStation,
					notifyType: "1",
					message:    "Ticket Conflict"}
				notify.notifyAdmin()
			}
		}
	}
	for ci:=0;ci<len(st.Cars);ci++{
		if st.Cars[ci].NoSeatCount != 0 {
			for ei:=0;ei<len(st.Cars[ci].EachRouteTravelerCount);ei++{
				if st.Cars[ci].EachRouteTravelerCount[ei] != copySt.Cars[ci].EachRouteTravelerCount[ei]{
					// TODO: 乘客人数不匹配，通知管理员处理
				}
			}
		}
		for si:=0;si<len(st.Cars[ci].Seats);si++{
			if st.Cars[ci].Seats[si].SeatBit != copySt.Cars[ci].Seats[si].SeatBit{
				// TODO：该座位存在问题，有可能是票冲突，也有可能是路段释放失败
			}
		}
	}
}

func getCarAndSeat(st *ScheduleTran, carNum uint8, seatType, seatNum string) (*ScheduleCar, *ScheduleSeat) {
	for ci := 0; ci < len(st.Cars); ci++ {
		if st.Cars[ci].CarNum == carNum {
			if seatType != constSeatTypeNoSeat {
				for si := 0; si < len(st.Cars[ci].Seats); si++ {
					if st.Cars[ci].Seats[si].SeatNum == seatNum {
						return &st.Cars[ci], &st.Cars[ci].Seats[si]
					}
				}
			}
			return &st.Cars[ci], nil
		}
	}
	return nil, nil
}
