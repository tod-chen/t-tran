package modules

import (
	"time"
)

var unpayOrders []order
var validOrders []order

type order struct{
	orderNum string
	userID int
	contactID int
	tranID int
	tranNum string
	carNum int
	seatNum string
	seatType string
	departureStation string
	departureTime time.Time
	arrivalStation string
	arrivalTime time.Time
	price float32
	bookTime time.Time
	payTime time.Time
	payType int
	payAccount string
	// 1.未支付 2.超时未支付 3.已支付 4.已支付并提醒
	// 5.已退票 6.已退票并提醒 7.已改签 8.已改签并提醒 9.已出票
	status int8
}


// SubmitOrder 订票
func SubmitOrder(t *tran, depIndex, arrIndex uint, userID, contactID int, isAdult bool, seatTypeIndex uint, acceptNoSeat bool)(msg string, ok bool){
	ok = false
	// 判断是否有未支付的订单，有未支付的订单，则不进行下一步 可考虑从数据库中查询
	length := len(unpayOrders)
	for i:=0;i<length;i++{
		if unpayOrders[i].userID == userID{
			msg = "您有未完成的订单，请先完成订单"
			return
		}
	}

	// 判断当前乘坐人在乘车时间上是否冲突 可考虑从数据库中查询
	length = len(validOrders)
	depTime := t.routeTimetable[depIndex].depTime
	arrTime := t.routeTimetable[arrIndex].arrTime
	for i:=0;i<length;i++{
		if validOrders[i].contactID == contactID{
			if (validOrders[i].arrivalTime.After(depTime) && validOrders[i].arrivalTime.Before(arrTime)) || 
			(validOrders[i].departureTime.After(depTime) && validOrders[i].departureTime.Before(arrTime)) || 
			(depTime.After(validOrders[i].departureTime) && depTime.Before(validOrders[i].arrivalTime)) {
				msg = "乘车人时间冲突"
				return
			}
		}
	}

	// 锁定座位，创建订单
	start, end := t.seatTypesIndex[seatTypeIndex], t.seatTypesIndex[seatTypeIndex + 1]
	seatMatch := getSeatMatch(depIndex, arrIndex)
	var s *seat

	for i:=start; i< end; i++ {
		if t.seats[i].IsAvailable(seatMatch, isAdult) {
			s = t.seats[i]
			break
		}
	}
	if acceptNoSeat && s == nil {
		typeIndex := uint(len(t.seatTypesIndex) - 2)
		var seatableTypeIndex uint
		for i:=typeIndex; i>0; i-- {
			if t.seatTypesIndex[i] != t.seatTypesIndex[i-1] {
				seatableTypeIndex = i - 1
				break
			}
		}
		if seatableTypeIndex == seatTypeIndex { // 可替换的座位类型
			s = t.getOneAvailableNoSeat(seatMatch, depIndex, arrIndex)
		}
	}
	if s != nil {
		s.l.Lock()
		if !s.isPutTogetherNoSeat {
			s.seatRoute ^= seatMatch
		} else{
			for i:=depIndex; i<arrIndex; i++ {
				t.noSeatEachRouteTravelerCount[i]++
			}
		}
		var o order
		o.bookTime = time.Now()
		o.departureStation = t.routeTimetable[depIndex].stationName
		o.departureTime = t.routeTimetable[depIndex].depTime
		o.arrivalStation = t.routeTimetable[arrIndex].stationName
		o.arrivalTime = t.routeTimetable[arrIndex].arrTime
		o.carNum = s.carNum
		o.contactID = contactID
		o.seatNum = s.seatNum
		o.userID = userID
		o.tranNum = t.tranNum
		o.tranID = t.id
		o.seatType = s.seatType
		o.price = getTicketPrice(t.tranNum, seatTypeIndex, depIndex, arrIndex) // 根据价格表，各路段累加
		unpayOrders = append(unpayOrders, o)
		s.l.Unlock()
		time.AfterFunc(unpayOrderAvaliableTime * time.Minute, func(){
			if o.status == 1 {
				o.status = 2
				s.l.Lock()
				if s.seatType != SeatTypeNoSeat {
					s.seatRoute &= (^seatMatch)
				} else {
					for i:=depIndex; i<arrIndex; i++ {
						t.noSeatEachRouteTravelerCount[i]--
					}
				}
				s.l.Unlock()
			}
		})
	}
	return
}

// Pay 订单支付
func (b *order)Pay(payType int, payAccount string, price float32) bool{
	if b.price != price{
		return false
	}
	b.payType = payType
	b.payAccount = payAccount
	b.payTime = time.Now()
	b.status = 3
	return true
}

// Refund 退票
func (b *order)Refund(){	
	b.status = 5
	for _, t := range AllTrans{
		if b.tranID != t.id{
			continue
		}
		for _, s := range t.seats{
			if b.seatNum == s.seatNum && b.carNum == s.carNum {
				depIndex, arrIndex := t.getDepAndArrIndexByStationName(b.departureStation, b.arrivalStation)
				if s.seatType != SeatTypeNoSeat {
					seatMatch := getSeatMatch(depIndex, arrIndex)
					s.l.Lock()
					s.seatRoute &= (^seatMatch)
					s.l.Unlock()
				} else {
					for i:=depIndex; i<arrIndex; i++ {
						t.noSeatEachRouteTravelerCount[i]++
					}
				}
				break
			}
		}
		break
	}
}

// Change 改签
func (b *order)Change(nb *order){
	
}

// CheckIn 取票
func (b *order)CheckIn(){
	b.status = 9
}