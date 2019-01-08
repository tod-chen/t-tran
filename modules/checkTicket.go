package modules

import "time"

// 已支付的全部有效订单
var allOrders []Order

// checkTicket 校验订单，以便释放无效订单所占用的资源 或 暴露冲突的订单
func checkTicket() {
	date := time.Now()
	for _, t := range tranInfos {
		for i := -1; i < constDays; i++ {
			dateStr := date.AddDate(0, 0, i).Format(ConstYmdFormat)
			checkTranTicket(t.TranNum, dateStr)
		}
	}
}

// checkTranTicket 校验具体某趟车次的订单
func checkTranTicket(tranNum, date string) {
	// 假设一趟车次总计有4000个有效订单
	list := make([]*Order, 4000)
	for i, o := range allOrders {
		if o.TranNum == tranNum && o.TranDepDate == date &&
			o.changeOrderID == 0 &&
			(o.Status == constOrderStatusPaid || o.Status == constOrderStatusIssued) {
			list = append(list, &allOrders[i])
		}
	}
	st := getScheduleTran(tranNum, date)
	t, _ := time.Parse(ConstYmdFormat, date)
	tran, exist := getTranInfo(tranNum, t)
	if !exist{
		return
	}
	for ci := 0; ci < len(st.Cars); ci++ {
		orderEachRouteTrvalerCount := make([]uint8, len(st.Cars[ci].EachRouteTravelerCount))
		for _, o := range list {
			if st.Cars[ci].CarNum == o.CarNum {
				for i := o.DepStationIdx; i < o.ArrStationIdx; i++ {
					orderEachRouteTrvalerCount[i]++
				}
			}
		}
		for i := 0; i < len(orderEachRouteTrvalerCount); i++ {
			// 当前车厢在当前路段，存在订单冲突
			if st.Cars[ci].EachRouteTravelerCount[i] < orderEachRouteTrvalerCount[i] {
				// 当前车厢在当前路段，多于席位数与站票数的总和
				if orderEachRouteTrvalerCount[i] > uint8(len(st.Cars[ci].Seats))+st.Cars[ci].NoSeatCount {
					notify := &notifyAdminInfo{
						date : t,
						tranNum: tranNum,
						carNum: st.Cars[ci].CarNum,
						depStation: tran.Timetable[i].StationName,
						arrStation: tran.Timetable[i + 1].StationName,
						notifyType: "1",
						message: "Ticket Conflict"}
					notify.notifyAdmin()
					// TODO: 暴露出来
				}
			} else if st.Cars[ci].EachRouteTravelerCount[i] > orderEachRouteTrvalerCount[i] {
				st.Cars[ci].EachRouteTravelerCount[i] = orderEachRouteTrvalerCount[i]
			}
		}
	}
}
