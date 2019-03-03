package modules

import (
	"errors"
	"time"
)

const (
	constOneDayOrderCancelLimit = 3 // 单日订单取消上限数

	// 订单状态
	constOrderUnpay     = iota // 未支付
	constOrderCancelled        // 取消未支付订单
	constOrderTimeout          // 未支付超时
	constOrderPaid             // 已支付
	constOrderRefund           // 已退款
	constOrderChanged          // 已改签

	// 车票状态
	constTicketUnpay        = iota // 未支付
	constTicketPaid                // 已支付
	constTicketRefund              // 已退票
	constTicketIssued              // 已出票
	constTicketChanged             // 已改签
	constTicketChangeUnpay         // 改签票未支付
	constTicketChangePaid          // 改签票已支付
	constTicketChangeRefund        // 改签票已退票
	constTicketChangeIssued        // 改签票已出票
	constTicketExpired             // 已过期（旅程已结束）
)

var (
	// 未支付订单
	unpayOrders []*Order
	// 已支付且未乘车的订单
	validOrders []*Order
)

// OrderCancelTimes 订单取消次数模型
type OrderCancelTimes struct {
	UserID      uint64    // 用户ID
	Date        time.Time // 日期
	CancelTimes uint8     // 取消订单的次数
}

// 判断用户当天取消订单的次数是否已达上限
func isCancelTimesInLimit(userID uint64) bool {
	date := time.Now().Format(ConstYmdFormat)
	oct := &OrderCancelTimes{}
	db.Where("user_id = ? and date = ?", userID, date).Attrs(OrderCancelTimes{CancelTimes: 0}).FirstOrCreate(oct)
	return oct.CancelTimes < constOneDayOrderCancelLimit
}

// Order 订单
type Order struct {
	ID            uint64
	OrderNum      string    // 订单号
	UserID        uint64    // 用户ID
	Price         float32   // 价格
	BookTime      time.Time `gorm:"type:datetime"` // 订票时间
	PayTime       time.Time `gorm:"type:datetime"` // 支付时间
	PayType       uint      // 支付类型
	PayAccount    string    // 支付账户
	Status        uint8     // 订单状态 0.未支付 1.已取消 2.订单超时 3.已支付 4.已退票 5.已改签
	ChangeOrderID uint64    // 改签票订单ID
}

// GetOrderInfo 获取订单信息
func GetOrderInfo(orderID uint64) *Order{
	order := Order{}
	db.Where("id = ?", orderID).First(&order)
	return &order
}

// hasUnpayOrder 判断是否有未支付订单
func hasUnpayOrder(userID uint64) bool {
	count := 0
	db.Model(&Order{}).Where("user_id = ? and status = 0", userID).Count(&count)
	return count == 0
}

func submitOrderValid(userID uint64, tranNum, date string) (*TranInfo, error) {
	dt, err := time.Parse(ConstYmdFormat, date)
	if err != nil {
		return nil, errors.New("日期无效")
	}
	if !isCancelTimesInLimit(userID) {
		return nil, errors.New("您单日取消订单次数已达上限")
	}
	if hasUnpayOrder(userID) {
		return nil, errors.New("您有未完成的订单，请先完成订单")
	}
	// 车次配置信息
	tran, exist := getTranInfo(tranNum, dt)
	if !exist {
		return nil, errors.New("车次信息不存在")
	}
	return tran, nil
}

// SubmitOrderModel 提交订单的请求结构体
type SubmitOrderModel struct{
	UserID uint64 `bson:"userID"`
	TranNum string `bson:"tranNum"`
	Date string `bson:"date"`
	DepIdx uint8 `bson:"depIdx"`
	ArrIdx uint8 `bson:"arrIdx"`
	PassengerIDs []uint64 `bson:"passengerIDs"`
	IsPortion bool `bson:"isPortion"`
	IsStudent bool 	`bson:"isStudent"`
	SeatType string `bson:"seatType"`
}

// SubmitOrder 订票
func SubmitOrder(par SubmitOrderModel) error {
	tran, err := submitOrderValid(par.UserID, par.TranNum, par.Date)
	if err != nil {
		return err
	}
	// 排班信息
	scheduleTran := getScheduleTran(par.TranNum, par.Date)
	carIdxList, exist := scheduleTran.carTypeIdxMap[par.SeatType]
	if !exist {
		return errors.New("所选席别无效")
	}
	pLen := len(par.PassengerIDs)
	if pLen == 1 {
		par.IsPortion = false
	}
	tickets, cars, seats := make([]*Ticket, 0, pLen), make([]*ScheduleCar, 0, pLen), make([]*ScheduleSeat, 0, pLen)
	// 获取发车时间和到站时间
	depTime, arrTime := tran.getDepAndArrTime(par.Date, par.DepIdx, par.ArrIdx)
	seatBit := countSeatBit(par.DepIdx, par.ArrIdx)
	for i := 0; i < pLen; i++ {
		if hasTimeConflict(par.PassengerIDs[i], depTime, arrTime) {
			return errors.New("乘车人时间冲突")
		}
		car, seat, isMedley, ok := bookSeat(scheduleTran, carIdxList, par.DepIdx, par.ArrIdx, seatBit, par.IsStudent)
		if ok {
			cars = append(cars, car)
			seats = append(seats, seat)
		} else if !par.IsPortion {
			// 无票 且要求全部提交时，释放已占用的资源 直接返回
			for i := 0; i < len(cars); i++ {
				seats[i].Release(seatBit)
				cars[i].releaseSeat(par.DepIdx, par.ArrIdx)
			}
			return errors.New("没有足够的票")
		}
		tickets = append(tickets, buildTicket(tran, car, seat, par.PassengerIDs[i], par.Date, par.DepIdx, par.ArrIdx, depTime, arrTime, isMedley, par.IsStudent))
	}
	o := &Order{
		ID:       getOrderID(par.UserID),
		OrderNum: "", // TODO: 订单号生成器需返回一个全局唯一订单号
		UserID:   par.UserID,
		BookTime: time.Now(),
		Status:   constOrderUnpay,
	}
	ticketIDs := getMultiTicketID(par.PassengerIDs)
	for i := 0; i < len(tickets); i++ {
		tickets[i].ID = ticketIDs[i]
		tickets[i].OrderID = o.ID
		db.Create(tickets[i])
		o.Price += tickets[i].Price
	}
	db.Create(o)
	// TODO：将新订单ID发向mq
	return nil
}

// 订座位
func bookSeat(st *ScheduleTran, carIdxList []uint8, depIdx, arrIdx uint8, seatBit int64, isStudent bool) (car *ScheduleCar, seat *ScheduleSeat, isMedley, ok bool) {
	// 优先席位票
	for _, carIdx := range carIdxList {
		if seat, ok = st.Cars[carIdx].getAvailableSeat(depIdx, arrIdx, seatBit, isStudent); ok {
			car = &st.Cars[carIdx]
			break
		}
	}
	// 无席位票，则考虑站票
	if !ok {
		for _, carIdx := range carIdxList {
			if seat, ok = st.Cars[carIdx].getAvailableNoSeat(depIdx, arrIdx, seatBit); ok {
				car = &st.Cars[carIdx]
				isMedley = true
				break
			}
		}
	}
	return
}

// ChangeOrder 改签
func ChangeOrder(par SubmitOrderModel, oldTicketID uint64) error {
	tran, err := submitOrderValid(par.UserID, par.TranNum, par.Date)
	if err != nil {
		return err
	}
	oldTicket := &Ticket{ID: oldTicketID}
	db.First(oldTicket)
	if oldTicket.ChangeTicketID != 0 {
		return errors.New("已经改签，无法再次改签")
	}
	// 出发时间和到站时间
	depTime, arrTime := tran.getDepAndArrTime(par.Date, par.DepIdx, par.ArrIdx)
	if hasTimeConflictInChange(par.PassengerIDs[0], oldTicketID, depTime, arrTime) {
		return errors.New("乘车人时间冲突")
	}
	scheduleTran := getScheduleTran(par.TranNum, par.Date)
	// 锁定座位，创建订单
	carIdxList, exist := scheduleTran.carTypeIdxMap[par.SeatType]
	if !exist {
		return errors.New("所选席别无效")
	}
	seatBit := countSeatBit(par.DepIdx, par.ArrIdx)
	car, seat, isMedley, ok := bookSeat(scheduleTran, carIdxList, par.DepIdx, par.ArrIdx, seatBit, par.IsStudent)
	// 无票
	if !ok {
		return errors.New("没有足够的票")
	}
	newTicket := buildTicket(tran, car, seat, par.PassengerIDs[0], par.Date, par.DepIdx, par.ArrIdx, depTime, arrTime, isMedley, par.IsStudent)
	newTicket.ID = getTicketID(par.PassengerIDs[0])
	newTicket.ChangeTicketID = oldTicket.ID
	newOrder := &Order{
		ID:       getOrderID(par.UserID),
		OrderNum: "", // TODO: 订单号生成器需返回一个全局唯一订单号
		UserID:   par.UserID,
		Price:    newTicket.Price,
		BookTime: time.Now(),
		Status:   constOrderUnpay,
	}
	oldOrder := &Order{ID: oldTicket.OrderID}
	db.First(oldOrder)
	oldOrder.ChangeOrderID = newOrder.ID
	// 原票价高于改签后的票价则需设置改签票为已支付状态，且需退还差额；否则改签票保持未支付状态，且用户需补交差额
	if oldOrder.Price >= newOrder.Price {
		// 退款
		Refund(oldOrder.ID, oldOrder.UserID, oldOrder.PayType, oldOrder.PayAccount, oldOrder.Price-newOrder.Price, time.Now().Format(ConstYMdHmsFormat))
		newOrder.Status = constOrderPaid
		oldOrder.Status = constOrderChanged
	} else {
		newOrder.Price -= oldOrder.Price
		// TODO：新订单ID需发向mq
	}
	return nil
}

// CancelOrder 取消订单
func CancelOrder(orderID uint64) error {
	o := &Order{ID: orderID}
	db.First(o)
	var tickets []Ticket
	db.Where("order_id = ?", o.ID).Find(&tickets)
	st := getScheduleTran(tickets[0].TranNum, tickets[0].TranDepDate)
	seatBit := countSeatBit(tickets[0].DepStationIdx, tickets[0].ArrStationIdx)
	for ti := 0; ti < len(tickets); ti++ {
		for ci := 0; ci < len(st.Cars); ci++ {
			if tickets[ti].CarNum == st.Cars[ci].CarNum {
				// 非站票需要释放资源
				if tickets[ti].SeatType != constSeatTypeNoSeat {
					for si := 0; si < len(st.Cars[ci].Seats); si++ {
						if tickets[ti].SeatNum == st.Cars[ci].Seats[si].SeatNum {
							st.Cars[ci].Seats[si].Release(seatBit)
							break
						}
					}
				}
				st.Cars[ci].releaseSeat(tickets[ti].DepStationIdx, tickets[ti].ArrStationIdx)
				break
			}
		}
	}
	o.Status = constOrderCancelled
	db.Save(o)
	return nil
}

// Payment 订单支付
func (o *Order) Payment(payType uint, payAccount string, price float32) error {
	if o.Status != constOrderUnpay {
		switch o.Status {
		case constOrderPaid:
			return errors.New("订单已支付")
		case constOrderTimeout:
			return errors.New("订单已过期")
		case constOrderRefund:
			return errors.New("订单已退款")
		case constOrderChanged:
			return errors.New("订单已改签")
		}
	}
	if o.Price != price {
		return errors.New("支付金额错误")
	}
	o.PayType = payType
	o.PayAccount = payAccount
	o.PayTime = time.Now()
	o.Status = constOrderPaid
	db.Save(o)
	return nil
}

// RefundOrder 退票（已支付订单）
func RefundOrder(orderID uint64) error {
	o := &Order{ID:orderID}
	db.First(o)
	if err := CancelOrder(orderID); err != nil {
		return err
	}
	if err := Refund(o.ID, o.UserID, o.PayType, o.PayAccount, o.Price, time.Now().Format(ConstYMdHmsFormat)); err != nil {
		return err
	}
	return nil
}

// Ticket 车票
type Ticket struct {
	ID              uint64
	OrderID         uint64    // 订单ID
	PassengerID     uint64    // 乘客ID
	IsStudent       bool      // 是否为学生票
	Status          uint8     // 车票状态
	Price           float32   // 票价
	TranDepDate     string    // 列车发车日期
	TranNum         string    // 车次号
	CarNum          uint8     // 车厢号
	SeatNum         string    // 座位号
	SeatType        string    // 座位类型
	CheckTicketGate string    // 检票口
	DepStation      string    // 出发站
	DepStationIdx   uint8     // 出发站在路线中的索引
	DepTime         time.Time `gorm:"type:datetime"` // 出发时间
	ArrStation      string    // 到达站
	ArrStationIdx   uint8     // 到达站在路线中的索引
	ArrTime         time.Time `gorm:"type:datetime"` // 到达时间
	ChangeTicketID  uint64    // 改签票的ID
}

// hasTimeConflict 判断乘车人的乘车时间是否冲突
func hasTimeConflict(passengerID uint64, depTime, arrTime time.Time) bool {
	count := 0
	validTicketStatus := []uint8{constTicketUnpay, constTicketPaid, constTicketIssued, constTicketChangeUnpay, constTicketChangePaid, constTicketChangeIssued}
	db.Model(&Ticket{}).Where("passenger_id = ? and status in (?) and ((dep_time < ? and ? < arr_time) or (dep_time < ? and ? < arr_time) or (? < dep_time and arr_time < ?))",
		passengerID, validTicketStatus, depTime, depTime, arrTime, arrTime, depTime, arrTime).Count(&count)
	return count != 0
}

// hasTimeConflictInChange 改签时，判断乘车人的乘车时间是否冲突
func hasTimeConflictInChange(passengerID, ticketID uint64, depTime, arrTime time.Time) bool {
	count := 0
	validTicketStatus := []uint8{constTicketUnpay, constTicketPaid, constTicketIssued, constTicketChangeUnpay, constTicketChangePaid, constTicketChangeIssued}
	// 相较于hasTimeConflict，多了一个ticketID的限制
	db.Model(&Ticket{}).Where("passenger_id = ? and status in (?) and id != ? and ((dep_time < ? and ? < arr_time) or (dep_time < ? and ? < arr_time) or (? < dep_time and arr_time < ?))",
		passengerID, validTicketStatus, ticketID, depTime, depTime, arrTime, arrTime, depTime, arrTime).Count(&count)
	return count != 0
}

// buildTicket 组装订单信息
func buildTicket(tran *TranInfo, car *ScheduleCar, seat *ScheduleSeat, passengerID uint64, date string, depIdx, arrIdx uint8, depTime, arrTime time.Time, isMedley, isStudent bool) *Ticket {
	t := &Ticket{
		PassengerID:     passengerID,
		IsStudent:       isStudent,
		Status:          constTicketUnpay,
		Price:           tran.getOrderPrice(car.SeatType, seat.SeatNum, depIdx, arrIdx),
		TranDepDate:     date,
		TranNum:         tran.TranNum,
		CarNum:          car.CarNum,
		SeatType:        car.SeatType,
		SeatNum:         seat.SeatNum,
		CheckTicketGate: tran.Timetable[depIdx].CheckTicketGate,
		DepStation:      tran.Timetable[depIdx].StationName,
		DepStationIdx:   depIdx,
		DepTime:         depTime,
		ArrStation:      tran.Timetable[arrIdx].StationName,
		ArrStationIdx:   arrIdx,
		ArrTime:         arrTime,
	}
	if isMedley {
		t.SeatType = constSeatTypeNoSeat
		t.SeatNum = ""
	}
	return t
}

// CheckIn 取票
func CheckIn(ticketID uint64) {
	t := &Ticket{ID: ticketID}
	db.First(t)
	if t.Status == constTicketPaid || t.Status == constTicketChangePaid {
		t.Status = constTicketIssued
	}
	db.Save(t)
}
