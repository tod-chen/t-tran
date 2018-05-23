package modules

import (
	"errors"
	"strings"
	"time"
)

const (
	// 订单状态：未支付、未支付超时、已支付、已改签、已出票
	constOrderStatusUnpay   = 1
	constOrderStatusTimeout = 2
	constOrderStatusPaid    = 3
	constOrderStatusChanged = 4
	constOrderStatusIssued  = 5
)

var (
	// 未支付订单
	unpayOrders []*order
	// 已支付且未乘车的订单
	validOrders []*order
)

func hasUnpayOrder(userID int) bool {
	length := len(unpayOrders)
	for i := 0; i < length; i++ {
		if unpayOrders[i].UserID == userID {
			return true
		}
	}
	return false
}

func hasTimeConflict(contactID int, depTime, arrTime time.Time) bool {
	length := len(validOrders)
	for i := 0; i < length; i++ {
		if validOrders[i].ContactID == contactID {
			if (validOrders[i].ArrivalTime.After(depTime) && validOrders[i].ArrivalTime.Before(arrTime)) ||
				(validOrders[i].DepartureTime.After(depTime) && validOrders[i].DepartureTime.Before(arrTime)) ||
				(depTime.After(validOrders[i].DepartureTime) && depTime.Before(validOrders[i].ArrivalTime)) {
				return true
			}
		}
	}
	return false
}

func hasTimeConflictForChangeOrder(contactID int, depTime, arrTime time.Time, oldOrder *order) bool {
	length := len(validOrders)
	for i := 0; i < length; i++ {
		if validOrders[i].ContactID == contactID && validOrders[i].OrderNum != oldOrder.OrderNum {
			if (validOrders[i].ArrivalTime.After(depTime) && validOrders[i].ArrivalTime.Before(arrTime)) ||
				(validOrders[i].DepartureTime.After(depTime) && validOrders[i].DepartureTime.Before(arrTime)) ||
				(depTime.After(validOrders[i].DepartureTime) && depTime.Before(validOrders[i].ArrivalTime)) {
				return true
			}
		}
	}
	return false
}

type order struct {
	OrderNum         string    `bson:"orderNum"`         // 订单号
	UserID           int       `bson:"userID"`           // 用户ID
	ContactID        int       `bson:"contactID"`        // 联系人ID
	TranDepDate      string    `bson:"tranDepDate"`      // 列车发车日期
	TranNum          string    `bson:"tranNum"`          // 车次号
	CarNum           uint8     `bson:"carNum"`           // 车厢号
	SeatNum          string    `bson:"seatNum"`          // 座位号
	SeatType         string    `bson:"seatType"`         // 座位类型
	DepartureStation string    `bson:"departureStation"` // 出发站
	CheckTicketGate  string    `bson:"checkTicketGate"`  // 检票口
	DepartureTime    time.Time `bson:"departureTime"`    // 出发时间
	ArrivalStation   string    `bson:"arrivalStation"`   // 到达站
	ArrivalTime      time.Time `bson:"arrivalTime"`      // 到达时间
	Price            float32   `bson:"price"`            // 票价
	BookTime         time.Time `bson:"bookTime"`         // 订票时间
	PayTime          time.Time `bson:"payTime"`          // 支付时间
	PayType          int       `bson:"payType"`          // 支付类型
	PayAccount       string    `bson:"payAccount"`       // 支付账户
	// 1.未支付 2.超时未支付 3.已支付 4.已退票 5.已改签 6.已出票
	Status int8 `bson:"status"` // 订单状态
	// 改签的票
	chargeOrder *order
}

func getOrderPrice(tranNum, seatType, seatNum string, depIdx, arrIdx uint8) (price float32) {
	var priceSlice []float32
	for i := 0; ; i++ {
		if tranList[i].TranNum == tranNum {
			switch seatType {
			case constSeatTypeAdvancedSoftSleeper, constSeatTypeSoftSleeper, constSeatTypeHardSleeper:
				priceSlice = tranList[i].SeatPriceMap[seatType+strings.Split(seatNum, "-")[1]]
			default:
				priceSlice = tranList[i].SeatPriceMap[seatType]
			}
			for j := depIdx; j < arrIdx; j++ {
				price += priceSlice[j]
			}
			break
		}
	}
	return 0
}

// SubmitOrder 订票
func SubmitOrder(t *Tran, depIdx, arrIdx uint8, userID, contactID int, isStudent bool, seatType string, acceptNoSeat bool) error {
	// 判断是否有未支付的订单，有未支付的订单，则不进行下一步 可考虑从数据库中查询
	if hasUnpayOrder(userID) {
		return errors.New("您有未完成的订单，请先完成订单")
	}

	// 判断当前乘坐人在乘车时间上是否冲突 可考虑从数据库中查询
	depTime := t.RouteTimetable[depIdx].DepTime
	arrTime := t.RouteTimetable[arrIdx].ArrTime
	if hasTimeConflict(contactID, depTime, arrTime) {
		return errors.New("乘车人时间冲突")
	}

	// 锁定座位，创建订单
	seatBit := countSeatBit(depIdx, arrIdx)
	carLen := len(t.Cars)
	var s *Seat
	success, isMedley, carIdx := false, false, -1
	for i := 0; i < carLen; i++ {
		if t.Cars[i].SeatType == seatType {
			// 优先查询非站票
			s, success = t.Cars[i].getAvailableSeat(seatBit, isStudent)
			if success {
				if isMedley {
					if t.Cars[i].occupySeat(depIdx, arrIdx) {
						carIdx = i
						break
					}
				} else if s.Book(seatBit, t.FullSeatBit, isStudent) {
					t.Cars[i].occupySeat(depIdx, arrIdx)
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
		for i := 0; i < carLen; i++ {
			if t.Cars[i].SeatType == seatType {
				s, success = t.Cars[i].getAvailableNoSeat(seatBit, depIdx, arrIdx)
				if success {
					if isMedley {
						if t.Cars[i].occupySeat(depIdx, arrIdx) {
							carIdx = i
							break
						}
					} else if s.Book(seatBit, t.FullSeatBit, true) {
						t.Cars[i].occupySeat(depIdx, arrIdx)
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
		BookTime:         time.Now(),
		DepartureStation: t.RouteTimetable[depIdx].StationName,
		DepartureTime:    t.RouteTimetable[depIdx].DepTime,
		ArrivalStation:   t.RouteTimetable[arrIdx].StationName,
		ArrivalTime:      t.RouteTimetable[arrIdx].ArrTime,
		CarNum:           t.Cars[carIdx].CarNum,
		ContactID:        contactID,
		SeatNum:          s.SeatNum,
		UserID:           userID,
		TranNum:          t.TranNum,
		TranDepDate:      t.DepartureDate,
		SeatType:         t.Cars[carIdx].SeatType,
		// 根据价格表，各路段累加
		Price: getOrderPrice(t.TranNum, t.Cars[carIdx].SeatType, s.SeatNum, depIdx, arrIdx),
	}
	// 拼凑的只能是站票
	if isMedley {
		o.SeatType = constSeatTypeNoSeat
	}
	unpayOrders = append(unpayOrders, o)
	time.AfterFunc(constUnpayOrderAvaliableTime*time.Minute, func() {
		if o.Status == constOrderStatusUnpay {
			o.Status = constOrderStatusTimeout
			s.Release(seatBit)
			t.Cars[carIdx].releaseSeat(depIdx, arrIdx)
		}
	})
	return nil
}

// CancelOrder 取消订单
func (o *order) CancelOrder() error {
	o.releaseSeat()
	//oy, om, od := o.departureTime.Date()
	return nil
}

// Pay 订单支付
func (o *order) Pay(payType int, payAccount string, price float32) error {
	switch o.Status {
	case constOrderStatusPaid:
		return errors.New("订单已支付")
	case constOrderStatusTimeout:
		return errors.New("订单已过期")
	case constOrderStatusChanged:
		return errors.New("订单已改签")
	case constOrderStatusIssued:
		return errors.New("已出票")
	}
	if o.Price != price {
		return errors.New("支付金额错误")
	}
	o.PayType = payType
	o.PayAccount = payAccount
	o.PayTime = time.Now()
	o.Status = constOrderStatusPaid
	// TODO：将此订单从未支付列表移至已支付列表
	return nil
}

// Refund 退票（已支付订单）
func (o *order) Refund() error {
	o.releaseSeat()
	return o.refundMoney(o.Price)
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
func (o *order) Change(t *Tran, depIndex, arrIndex uint, userID, contactID int, isStudent bool, seatTypeIndex uint, acceptNoSeat bool) (msg string, ok bool) {
	ok = false
	// 判断是否有未支付的订单，有未支付的订单，则不进行下一步 可考虑从数据库中查询
	if hasUnpayOrder(userID) {
		msg = "您有未完成的订单，请先完成订单"
		return
	}

	// 判断当前乘坐人在乘车时间上是否冲突 可考虑从数据库中查询
	depTime := t.RouteTimetable[depIndex].DepTime
	arrTime := t.RouteTimetable[arrIndex].ArrTime
	if hasTimeConflictForChangeOrder(contactID, depTime, arrTime, o) {
		msg = "乘车人时间冲突"
		return
	}

	// 锁定座位，创建订单
	return
}

// CheckIn 取票
func (o *order) CheckIn() {
	if o.Status == constOrderStatusPaid {
		o.Status = constOrderStatusIssued
	}
}

func (o *order) releaseSeat() error {
	return nil
}

func (o *order) refundMoney(amount float32) error {
	return nil
}
