package modules

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 初始化列车排班
func initTranScheduleMap() {
	fmt.Println("init scheduleTranMap beginning")
	defer fmt.Println("init scheduleTranMap end")

	scheduleTranMap = make(map[string]([]ScheduleTran), 30)
	now := time.Now()
	for i := 0; i < len(tranList); i++ {
		fmt.Println("init schedule tran: ", tranList[i].TranNum)
		// 可提前30天订票，但是要准备好40天的排班，留给后台操作的时间为不可订票的10天
		for d := 0; d < constDays+10; d++ {
			initScheduleByDate(now, d, &tranList[i])
		}
		// 对于全程运行时间超过一天的，如果当前时间之前的排班也需初始化
		if tranList[i].DurationDays > 1 {
			for d := 1 - tranList[i].DurationDays; d < 0; d++ {
				initScheduleByDate(now, int(d), &tranList[i])
			}
		}
	}
}

func initScheduleByDate(now time.Time, day int, t *TranInfo) {
	date, count := now.AddDate(0, 0, day), 0
	dFormat := date.Format(constYmdFormat)
	sTran := ScheduleTran{}
	q := db.Table("schedule_trans").Where("departure_date = ? and tran_num = ?", dFormat, t.TranNum)
	q.Count(&count)
	if count == 0 {
		sTran = ScheduleTran{DepartureDate: dFormat, TranNum: t.TranNum}
		// tranList 中相同的车次已经按 EnalbeLevel 排序好了，优先排等级较高的一个
		if date.After(t.EnableStartDate) &&
			date.Before(t.EnableEndDate) {
			// 对于不是每天发车的车次，需要根据其间隔发车天数来初始化
			if t.DurationDays > 1 {
				durFormat := date.AddDate(0, 0, 1-int(t.DurationDays)).Format(constYmdFormat)
				db.Table("schedule_trans").Where("tran_num = ? and strcmp(departure_date, ?) <= 0 and strcmp(departure_date, ?) >= 0", dFormat, durFormat).Count(&count)
				// 站发车间隔时间内，已经有该车次的排班数据，则跳过
				if count != 0 {
					return
				}
			}
			initScheduleTran(&sTran, t, false)
		}
	} else {
		q.First(&sTran)
		initScheduleTran(&sTran, t, true)
	}
	trans := scheduleTranMap[dFormat]
	trans = append(trans, sTran)
	scheduleTranMap[dFormat] = trans
}

func initScheduleTran(st *ScheduleTran, t *TranInfo, exist bool) {
	routeCount := uint8(len(t.RouteTimetable) - 1)
	st.FullSeatBit = countSeatBit(0, routeCount)
	if !exist {
		db.Create(st)
		scheduleCars := make([]ScheduleCar, len(t.Cars))
		for i := 0; i < len(t.Cars); i++ {
			scheduleCars[i] = ScheduleCar{
				ScheduleTranID: st.ID,
				CarNum:         uint8(i + 1),
				NoSeatCount:    t.Cars[i].NoSeatCount,
			}
			// 站票数不为零，才初始化EachRouteTravelerCount
			if t.Cars[i].NoSeatCount != 0 {
				tempStrArr := make([]string, routeCount)
				for j := 0; j < len(tempStrArr); j++ {
					tempStrArr[j] = "0"
				}
				scheduleCars[i].EachRouteTravelerCount = make([]uint8, routeCount)
				scheduleCars[i].EachRouteTravelerCountStr = strings.Join(tempStrArr, ",")
			}
			db.Create(&scheduleCars[i])
			var seats []Seat
			db.Where("car_id = ?", t.Cars[i].ID).Order("seat_num").Find(&seats)
			scheduleSeats := make([]ScheduleSeat, len(seats))
			for j := 0; j < len(seats); j++ {
				scheduleSeats[j].CarID = scheduleCars[i].ID
				scheduleSeats[j].SeatNum = seats[j].SeatNum
				scheduleSeats[j].IsStudent = seats[j].IsStudent
				db.Create(&scheduleSeats[j])
			}
			scheduleCars[i].Seats = scheduleSeats
		}
	} else {
		db.Save(st)
		var cars []ScheduleCar
		db.Where("schedule_tran_id", st.ID).Order("car_num").Find(&cars)
		//db.Where("schedule_tran_id = ?", st.ID).Find(&cars)
		st.Cars = cars
		for i := 0; i < len(st.Cars); i++ {
			if st.Cars[i].NoSeatCount != 0 {
				tempStrArr := strings.Split(st.Cars[i].EachRouteTravelerCountStr, ",")
				st.Cars[i].EachRouteTravelerCount = make([]uint8, len(tempStrArr))
				for j := 0; j < len(tempStrArr); j++ {
					count, _ := strconv.Atoi(tempStrArr[j])
					st.Cars[i].EachRouteTravelerCount[j] = uint8(count)
				}
			}
			var seats []ScheduleSeat
			db.Where("car_id", st.Cars[i].ID).Order("seat_num").Find(&seats)
			st.Cars[i].Seats = seats
		}
	}
}

// ScheduleTran 车次排班信息
type ScheduleTran struct {
	DepartureDate  string        `gorm:"index:q;type:varchar(10)" json:"departureDate"` // 发车日期
	TranNum        string        `gorm:"index:q;type:varchar(10)" json:"tranNum"`       // 车次号
	SaleTicketTime time.Time     `gorm:"type:datetime" json:"saleTicketTime"`           // 售票时间
	NotSaleRemark  string        `json:"notSaleRemark"`                                 // 不售票的说明，路线调整啥的
	Cars           []ScheduleCar `gorm:"foreignkey:ScheduleTranID"`                     // 车厢
	// 各种类型车厢的起始索引
	// 依次是：商务座、一等座、二等座、高级软卧、软卧、动卧、硬卧、软座、硬座
	// 若值是[0, 1, 4, 6, 6, 6, 6, 6, 10, 15]，
	// 表示第一个车厢是商务座; 第2-4个车厢是一等座; 第5-6个车厢是二等座; 高级软卧、软卧、动卧、硬卧都没有; 第7-10个是软座; 第11-15个是硬座;
	carTypesIdx []uint8
	FullSeatBit uint64 // 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
	DBModel
}

// ScheduleCar 排班中的车厢
type ScheduleCar struct {
	ScheduleTranID            uint           // 车次排班ID
	CarNum                    uint8          // 车厢号
	NoSeatCount               uint8          // 车厢内站票数
	Seats                     []ScheduleSeat `gorm:"foreignkey:CarID"` // 车厢的所有座位
	EachRouteTravelerCountStr string         // 各路段乘客人数，用英文逗号分隔存于数据库
	// 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	EachRouteTravelerCount []uint8 `gorm:"-"`
	sync.RWMutex                   // 读写锁，用于保护各路段乘客人数字段
	DBModel
}

// ScheduleSeat 排班中的座位
type ScheduleSeat struct {
	SeatBit    uint64 // 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	sync.Mutex        // 锁，订票与退票均需要锁
	Seat
}

// getAvailableSeat 获取可预订的座位,是否获取成功标记,是否为拼凑的站票标记
// func (c *Car) getAvailableSeat(seatBit uint64, isStudent bool) (s *Seat, ok bool) {
// 	for i := 0; i < len(c.Seats); i++ {
// 		if c.Seats[i].IsAvailable(seatBit, isStudent) {
// 			return &c.Seats[i], true
// 		}
// 	}
// 	return nil, false
// }

// func (c *Car) getAvailableNoSeat(seatBit uint64, depIdx, arrIdx uint8) (s *Seat, ok bool) {
// 	if c.NoSeatCount == 0 {
// 		return nil, false
// 	}
// 	// 下面开始查找拼凑的站票
// 	// 非站票数与站票数之和
// 	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
// 	// 旅途中当前车厢旅客最大数
// 	var maxTravelerCountInRoute uint8
// 	c.RLock()
// 	defer c.RUnlock()
// 	for i := depIdx; i < arrIdx; i++ {
// 		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
// 			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
// 		}
// 	}
// 	if totalSeatCount-int(maxTravelerCountInRoute) > 0 {
// 		s = &Seat{SeatNum: "", SeatBit: seatBit}
// 		return s, true
// 	}
// 	return nil, false
// }

// func (c *Car) getAvailableNoSeatCount(seatBit uint64, depIdx, arrIdx uint8) uint8 {
// 	// 车厢未设置站票时，直接返回 0
// 	if c.NoSeatCount == 0 {
// 		return 0
// 	}
// 	// 非站票数与站票数之和
// 	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
// 	// 旅途中当前车厢旅客最大数
// 	var maxTravelerCountInRoute uint8
// 	c.RLock()
// 	defer c.RUnlock()
// 	for i := depIdx; i < arrIdx; i++ {
// 		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
// 			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
// 		}
// 	}
// 	return uint8(totalSeatCount) - maxTravelerCountInRoute
// }

// 某座位被占用
// func (c *Car) occupySeat(depIdx, arrIdx uint8) bool {
// 	if c.NoSeatCount == 0 {
// 		return true
// 	}
// 	c.Lock()
// 	defer c.Unlock()
// 	var maxCount uint8
// 	for i := depIdx; i < arrIdx; i++ {
// 		if c.EachRouteTravelerCount[i] > maxCount {
// 			maxCount = c.EachRouteTravelerCount[i]
// 		}
// 	}
// 	if maxCount >= uint8(len(c.Seats))+c.NoSeatCount {
// 		return false
// 	}
// 	for i := depIdx; i < arrIdx; i++ {
// 		c.EachRouteTravelerCount[i]++
// 	}
// 	return true
// }

// 某座位被释放
// func (c *Car) releaseSeat(depIdx, arrIdx uint8) {
// 	if c.NoSeatCount != 0 {
// 		c.Lock()
// 		defer c.Unlock()
// 		for i := depIdx; i < arrIdx; i++ {
// 			c.EachRouteTravelerCount[i]--
// 		}
// 	}
// }

// IsAvailable 根据路段和乘客类型判断能否订票
// func (s *SeatSchedule) IsAvailable(seatBit uint64, isStudent bool) bool {
// 	return (s.IsStudent == isStudent) && (s.SeatBit^seatBit == s.SeatBit+seatBit)
// }

// // Book 订票
// func (s *SeatSchedule) Book(seatBit, tranFullSeatBit uint64, isStudent bool) bool {
// 	s.Lock()
// 	defer s.Unlock()
// 	if !s.IsAvailable(seatBit, isStudent) {
// 		return false
// 	}
// 	s.SeatBit ^= seatBit
// 	return true
// }

// // Release 退票或取消订单，释放座位对应路段的资源
// func (s *SeatSchedule) Release(seatBit uint64) {
// 	s.Lock()
// 	defer s.Unlock()
// 	s.SeatBit ^= (^seatBit)
// }
