package modules

import (
	"strings"
	"errors"
	"time"
)

const (
	// 订单状态：未支付、未支付超时、已支付、已改签、已出票
	constOrderStatusUnpay = 1
	constOrderStatusTimeout = 2
	constOrderStatusPaid = 3
	constOrderStatusChanged = 4
	constOrderStatusIssued = 5
)
var (
	// 未支付订单
	unpayOrders []*order
	// 已支付且未乘车的订单
	validOrders []*order
)

func hasUnpayOrder(userID int)bool{
	length := len(unpayOrders)
	for i:=0;i<length;i++{
		if unpayOrders[i].userID == userID{
			return true
		}
	}
	return false
}

func hasTimeConflict(contactID int, depTime, arrTime time.Time) bool{
	length := len(validOrders)
	for i:=0;i<length;i++{
		if validOrders[i].contactID == contactID{
			if (validOrders[i].arrivalTime.After(depTime) && validOrders[i].arrivalTime.Before(arrTime)) || 
			(validOrders[i].departureTime.After(depTime) && validOrders[i].departureTime.Before(arrTime)) || 
			(depTime.After(validOrders[i].departureTime) && depTime.Before(validOrders[i].arrivalTime)) {
				return true
			}
		}
	}
	return false
}

func hasTimeConflictForChangeOrder(contactID int, depTime, arrTime time.Time, oldOrder *order) bool{
	length := len(validOrders)
	for i:=0;i<length;i++{
		if validOrders[i].contactID == contactID && validOrders[i].orderNum != oldOrder.orderNum{
			if (validOrders[i].arrivalTime.After(depTime) && validOrders[i].arrivalTime.Before(arrTime)) || 
			(validOrders[i].departureTime.After(depTime) && validOrders[i].departureTime.Before(arrTime)) || 
			(depTime.After(validOrders[i].departureTime) && depTime.Before(validOrders[i].arrivalTime)) {
				return true
			}
		}
	}
	return false
}

type order struct{
	orderNum string
	userID int
	contactID int
	tranID int
	tranNum string
	carNum uint8
	seatNum string
	seatType string
	departureStation string
	checkTicketGate string
	departureTime time.Time
	arrivalStation string
	arrivalTime time.Time
	price float32
	bookTime time.Time
	payTime time.Time
	payType int
	payAccount string
	// 1.未支付 2.超时未支付 3.已支付 4.已退票 5.已改签 6.已出票
	status int8
	// 改签的票
	chargeOrder *order
}

func getOrderPrice(tranNum, seatType, seatNum string, depIdx, arrIdx uint8) (price float32) {
	var priceSlice []float32
	for i:=0; ; i++ {
		if tranList[i].tranNum == tranNum {
			switch seatType {
			case constSeatTypeAdvancedSoftSleeper, constSeatTypeSoftSleeper, constSeatTypeHardSleeper :
				priceSlice = tranList[i].seatPriceMap[seatType + strings.Split(seatNum, "-")[1]]
			default:
				priceSlice = tranList[i].seatPriceMap[seatType]
			}
			for j:=depIdx; j<arrIdx; j++ {
				price += priceSlice[j]
			}
			break
		}
	}
	return 0
}

// SubmitOrder 订票
func SubmitOrder(t *tran, depIdx, arrIdx uint8, userID, contactID int, isStudent bool, seatType string, acceptNoSeat bool)error{	
	// 判断是否有未支付的订单，有未支付的订单，则不进行下一步 可考虑从数据库中查询
	if hasUnpayOrder(userID) {
		return errors.New("您有未完成的订单，请先完成订单")
	}

	// 判断当前乘坐人在乘车时间上是否冲突 可考虑从数据库中查询
	depTime := t.routeTimetable[depIdx].depTime
	arrTime := t.routeTimetable[arrIdx].arrTime
	if hasTimeConflict(contactID, depTime, arrTime) {
		return errors.New("乘车人时间冲突")
	}

	// 锁定座位，创建订单
	seatBit := countSeatBit(depIdx, arrIdx)
	carLen := len(t.cars)
	var s *seat
	success, isMedley, carIdx := false, false, -1
	for i:=0; i<carLen; i++ {
		if t.cars[i].seatType == seatType {
			// 优先查询非站票
			s, success, isMedley = t.cars[i].getAvailableSeat(seatBit, isStudent, false, depIdx, arrIdx)
			if success {
				if isMedley {
					if t.cars[i].occupySeat(depIdx, arrIdx) {
						carIdx = i
						break
					}
				} else if s.Book(seatBit, t.fullSeatBit, isStudent) {
					t.cars[i].occupySeat(depIdx, arrIdx)
					carIdx = i
					break
				}
			}
		}
	}
	// 未订到非站票
	if !success {
		// 不接受站票的情况下，直接返回
		if !acceptNoSeat {
			return errors.New("没有足够的票")
		}
		// 继续预订站票
		for i:=0; i<carLen; i++ {
			if t.cars[i].seatType == seatType {
				s, success, isMedley = t.cars[i].getAvailableSeat(seatBit, true, true, depIdx, arrIdx)
				if success {
					if isMedley {
						if t.cars[i].occupySeat(depIdx, arrIdx) {
							carIdx = i
							break
						}
					} else if s.Book(seatBit, t.fullSeatBit, true) {
						t.cars[i].occupySeat(depIdx, arrIdx)
						carIdx = i
						break
					}
				}
			}
		}
	}
	if !success {
		return errors.New("没有足够的票")
	}
	o := &order{
		bookTime : time.Now(),
		departureStation : t.routeTimetable[depIdx].stationName,
		departureTime : t.routeTimetable[depIdx].depTime,
		arrivalStation : t.routeTimetable[arrIdx].stationName,
		arrivalTime : t.routeTimetable[arrIdx].arrTime,
		carNum : t.cars[carIdx].carNum,
		contactID : contactID,
		seatNum : s.seatNum,
		userID : userID,
		tranNum : t.tranNum,
		tranID : t.id,
		seatType : t.cars[carIdx].seatType,
		// 根据价格表，各路段累加
		price : getOrderPrice(t.tranNum, t.cars[carIdx].seatType, s.seatNum, depIdx, arrIdx),
	}
	// 拼凑的只能是站票
	if isMedley {
		o.seatType = constSeatTypeNoSeat
	}
	unpayOrders = append(unpayOrders, o)
	time.AfterFunc(constUnpayOrderAvaliableTime * time.Minute, func(){
		if o.status == constOrderStatusUnpay {
			o.status = constOrderStatusTimeout
			s.Release(seatBit)
			t.cars[carIdx].releaseSeat(depIdx, arrIdx)
		}
	})
	return nil
}

// CancelOrder 取消订单
func (o *order)CancelOrder()error{
	o.releaseSeat()
	//oy, om, od := o.departureTime.Date()
	return nil
}

// Pay 订单支付
func (o *order)Pay(payType int, payAccount string, price float32) error{
	switch o.status {
	case constOrderStatusPaid:
		return errors.New("订单已支付")
	case constOrderStatusTimeout:
		return errors.New("订单已过期")
	case constOrderStatusChanged:
		return errors.New("订单已改签")
	case constOrderStatusIssued:
		return errors.New("已出票")
	}
	if o.price != price{
		return errors.New("支付金额错误")
	}
	o.payType = payType
	o.payAccount = payAccount
	o.payTime = time.Now()
	o.status = constOrderStatusPaid
	// TODO：将此订单从未支付列表移至已支付列表
	return nil
}

// Refund 退票（已支付订单）
func (o *order)Refund()error{
	o.releaseSeat()
	return o.refundMoney(o.price)
	// for _, t := range allTrans{
	// 	if o.tranID != t.id{
	// 		continue
	// 	}
	// 	for _, s := range t.seats{
	// 		if o.seatNum == s.seatNum && o.carNum == s.carNum {
	// 			depIndex, arrIndex := t.getDepAndArrIndexByStationName(o.departureStation, o.arrivalStation)
	// 			if s.seatType != SeatTypeNoSeat {
	// 				seatMatch := getSeatMatch(depIndex, arrIndex)
	// 				s.l.Lock()
	// 				s.seatRoute &= (^seatMatch)
	// 				s.l.Unlock()
	// 			} else {
	// 				for i:=depIndex; i<arrIndex; i++ {
	// 					t.noSeatEachRouteTravelerCount[i]++
	// 				}
	// 			}
	// 			break
	// 		}
	// 	}
	// 	break
	// }
	// return nil
}

// Change 改签
func (o *order)Change(t *tran, depIndex, arrIndex uint, userID, contactID int, isStudent bool, seatTypeIndex uint, acceptNoSeat bool)(msg string, ok bool){
	ok = false
	// 判断是否有未支付的订单，有未支付的订单，则不进行下一步 可考虑从数据库中查询
	if hasUnpayOrder(userID) {
		msg = "您有未完成的订单，请先完成订单"
		return
	}

	// 判断当前乘坐人在乘车时间上是否冲突 可考虑从数据库中查询
	depTime := t.routeTimetable[depIndex].depTime
	arrTime := t.routeTimetable[arrIndex].arrTime
	if hasTimeConflictForChangeOrder(contactID, depTime, arrTime, o) {
		msg = "乘车人时间冲突"
		return
	}


	// 锁定座位，创建订单
	return
}

// CheckIn 取票
func (o *order)CheckIn(){
	if o.status == constOrderStatusPaid{
		o.status = constOrderStatusIssued
	}
}

func (o *order)releaseSeat() error{
	return nil
}

func (o *order)refundMoney(amount float32) error {
	return nil
}