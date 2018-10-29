package modules

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	constDays                    = 30  // 可提前订票天数
	constQueryTranDelay          = 20  // 查询时距离当前时间多长时间内发车的车次不予显示 单位：分钟
	constUnpayOrderAvaliableTime = 45  // 未完成订单有效时间 单位：分钟
	constMaxAvaliableSeatCount   = 100 // 查询余票数量的最大值 超过此值时，显示“有”
)

var (
	// 列车按日期安排表
	scheduleTranMap map[string]([]ScheduleTran)
)

// 初始化列车排班
func initTranSchedule() {
	fmt.Println("init scheduleTran beginning")
	defer fmt.Println("init scheduleTran end")

	scheduleTranMap = make(map[string]([]ScheduleTran), constDays)
	now := time.Now()
	for i := 0; i < len(tranInfos); i++ {
		fmt.Println("init schedule tran count: ", i)
		//fmt.Println("init schedule tran: ", tranInfos[i].TranNum)
		// 可提前30天订票，但是要准备好40天的排班，留给后台操作的时间为不可订票的10天
		for d := 0; d < constDays+10; d++ {
			initScheduleByDate(now.AddDate(0, 0, d), &tranInfos[i])
		}
		// 对于全程运行时间超过一天的，如果当前时间之前的排班也需初始化
		if tranInfos[i].DurationDays > 1 {
			for d := 1 - tranInfos[i].DurationDays; d < 0; d++ {
				initScheduleByDate(now.AddDate(0, 0, int(d)), &tranInfos[i])
			}
		}
	}
}

func initScheduleByDate(date time.Time, t *TranInfo) {
	strDate := date.Format(constYmdFormat)
	tranScheduleColl := getMgoDB().C("scheduleTran")
	sTran := ScheduleTran{}
	q := tranScheduleColl.Find(bson.M{"departureDate": strDate, "tranNum": t.TranNum})
	count, err := q.Count()
	if err != nil {
		panic(err)
	}
	if count != 0 {
		if err = q.One(&sTran); err != nil {
			panic(err)
		}
	} else {

	}
	q := db.Table("schedule_trans").Where("departure_date = ? and tran_num = ?", strDate, t.TranNum)
	q.Count(&count)
	if count == 0 {
		// tranList 中相同的车次已经按 EnalbeLevel 排序好了，优先排等级较高的一个
		sTran = ScheduleTran{DepartureDate: strDate, TranNum: t.TranNum}
		if date.After(t.EnableStartDate) &&
			date.Before(t.EnableEndDate) {
			// 对于不是每天发车的车次，需要根据其间隔发车天数来初始化
			if t.DurationDays > 1 {
				durFormat := date.AddDate(0, 0, 1-int(t.DurationDays)).Format(constYmdFormat)
				db.Table("schedule_trans").Where("tran_num = ? and strcmp(departure_date, ?) <= 0 and strcmp(departure_date, ?) >= 0", t.TranNum, strDate, durFormat).Count(&count)
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
	trans := scheduleTranMap[strDate]
	trans = append(trans, sTran)
	scheduleTranMap[strDate] = trans
}

func initScheduleTran(st *ScheduleTran, t *TranInfo, exist bool) {
	routeCount := uint8(len(t.Timetable) - 1)
	st.FullSeatBit = countSeatBit(0, routeCount)
	if !exist {
		db.Create(st)
		scheduleCars := make([]ScheduleCar, len(t.Cars))
		for i := 0; i < len(t.Cars); i++ {
			scheduleCars[i] = ScheduleCar{
				ScheduleTranID: st.ID,
				SeatType:       t.Cars[i].SeatType,
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
		db.Where("schedule_tran_id = ?", st.ID).Order("car_num").Find(&cars)
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
			db.Where("car_id = ?", st.Cars[i].ID).Order("seat_num").Find(&seats)
			st.Cars[i].Seats = seats
		}
	}
}

// 余票信息按出发时间排序，也可以丢到前端去排序，js的排序还是比较方便的
type residualTicketSort []*ResidualTicketInfo

func (r residualTicketSort) Len() int           { return len(r) }
func (r residualTicketSort) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r residualTicketSort) Less(i, j int) bool { return r[i].depTime.Before(r[j].depTime) }

// ResidualTicketInfo 余票信息结构
type ResidualTicketInfo struct {
	tranNum   string    // 车次号
	date      string    // 发车日期
	depIdx    uint8     // 出发站索引，值为0时表示出发站为起点站，否则表示路过
	depCode   string    // 出发站编码
	depTime   time.Time // 出发时间, 满足条件的列车需根据出发时间排序
	arrIdx    uint8     // 目的站索引，值与routeCount相等时表示目的站为终点，否则表示路过
	arrCode   string    // 目的站编码
	arrTime   string    // 到达时间
	costTime  string    // 历时，根据出发时间与历时可计算出跨天数，在前端计算即可
	seatCount []int     // 各座次余票数
	remark    string    // 不售票的说明
}

func newResidualTicketInfo(t *TranInfo, depIdx, arrIdx uint8) *ResidualTicketInfo {
	r := &ResidualTicketInfo{
		tranNum:  t.TranNum,
		date:     t.Timetable[0].DepTime.Format(constYmdFormat),
		depIdx:   depIdx,
		depCode:  t.Timetable[depIdx].StationCode,
		depTime:  t.Timetable[depIdx].DepTime,
		arrIdx:   arrIdx,
		arrCode:  t.Timetable[arrIdx].StationCode,
		arrTime:  t.Timetable[arrIdx].ArrTime.Format(constHmFormat),
		costTime: t.Timetable[arrIdx].ArrTime.Sub(t.Timetable[depIdx].DepTime).String(),
		// 初始化全部没有座位
		seatCount: []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
	}
	return r
}

func (r *ResidualTicketInfo) setScheduleInfo(st *ScheduleTran, isStudent bool) {
	r.remark = st.NotSaleRemark
	if seatCount, success := st.GetSeatCount(r.depIdx, r.arrIdx, isStudent); success {
		r.seatCount = seatCount
	}
}

// 结果转为字符串
func (r *ResidualTicketInfo) toString() string {
	list := []string{r.tranNum, r.date,
		strconv.Itoa(int(r.depIdx)), r.depCode, r.depTime.Format(constHmFormat),
		strconv.Itoa(int(r.arrIdx)), r.arrCode, r.arrTime, r.costTime}
	countList := make([]string, 12)
	count := 0
	for i := 0; i < len(r.seatCount); i++ {
		count = r.seatCount[i]
		if count == 0 {
			countList[i] = "无"
		}
		if 0 < count && count < constMaxAvaliableSeatCount {
			countList[i] = strconv.Itoa(count)
		}
		if count >= constMaxAvaliableSeatCount {
			countList[i] = "有"
		}
	}
	countList[11] = r.remark
	list = append(list, countList...)
	return strings.Join(list, "|")
}

// ScheduleTran 车次排班信息
type ScheduleTran struct {
	DepartureDate  string        `gorm:"index:q;type:varchar(10)" json:"departureDate"` // 发车日期
	TranNum        string        `gorm:"index:q;type:varchar(10)" json:"tranNum"`       // 车次号
	SaleTicketTime time.Time     `gorm:"type:datetime" json:"saleTicketTime"`           // 售票时间
	NotSaleRemark  string        `json:"notSaleRemark"`                                 // 不售票的说明，路线调整啥的
	Cars           []ScheduleCar `gorm:"-"`                                             // 车厢
	// 各种类型车厢的起始索引
	// 依次是：商务座、一等座、二等座、高级软卧、软卧、动卧、硬卧、软座、硬座
	// 若值是[0, 1, 4, 6, 6, 6, 6, 6, 10, 15]，
	// 表示第一个车厢是商务座; 第2-4个车厢是一等座; 第5-6个车厢是二等座; 高级软卧、软卧、动卧、硬卧都没有; 第7-10个是软座; 第11-15个是硬座;
	carTypesIdx []uint8 `gorm:"-"`
	FullSeatBit uint64  // 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
	DBModel
}

func getScheduleTran(tranNum, date string) *ScheduleTran {
	count := len(scheduleTranMap[date])
	idx := sort.Search(count, func(i int) bool {
		return scheduleTranMap[date][i].TranNum >= tranNum
	})
	return &scheduleTranMap[date][idx]
}

// Save 保存到数据库
func (st *ScheduleTran) Save() (bool, string) {
	st.SaleTicketTime = st.SaleTicketTime.Add(-8 * time.Hour)
	db.Save(st)
	return true, ""
}

// GetSeatCount 获取余票信息
func (st *ScheduleTran) GetSeatCount(depIdx, arrIdx uint8, isStudent bool) (result []int, success bool) {
	// 有不售票说明，表示当前不售票，不用查余票数
	if st.NotSaleRemark != "" {
		return
	}
	// 当前时间未开售，不用查询余票数
	if st.SaleTicketTime.After(time.Now()) {
		return
	}
	// 总共11类座次
	result = make([]int, 11)
	// 路段位标记
	seatBit := countSeatBit(depIdx, arrIdx)
	var noSeatCount uint8
	for i := 0; i < len(st.carTypesIdx)-1; i++ {
		start, end := st.carTypesIdx[i], st.carTypesIdx[i+1]
		if start == end {
			result[i] = -1
			continue
		}
		availableSeatCount := 0
		for j := start; j < end && availableSeatCount < constMaxAvaliableSeatCount; j++ {
			for k := 0; k < len(st.Cars[j].Seats) && availableSeatCount < constMaxAvaliableSeatCount; k++ {
				if st.Cars[j].Seats[k].IsAvailable(seatBit, isStudent) {
					availableSeatCount++
				}
			}
			if noSeatCount < constMaxAvaliableSeatCount {
				noSeatCount += st.Cars[j].getNoSeatCount(seatBit, depIdx, arrIdx)
			}
		}
		result[i] = availableSeatCount
	}
	result[len(result)-1] = int(noSeatCount)
	success = true
	return
}

// ScheduleCar 排班中的车厢
type ScheduleCar struct {
	ScheduleTranID uint64         // 车次排班ID
	SeatType       string         // 座次
	CarNum         uint8          // 车厢号
	NoSeatCount    uint8          // 车厢内站票数
	Seats          []ScheduleSeat // 车厢的所有座位
	//EachRouteTravelerCountStr string         // 各路段乘客人数，用英文逗号分隔存于数据库
	// 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	EachRouteTravelerCount []uint8 //`gorm:"-"`
	sync.RWMutex                   // 读写锁，用于保护各路段乘客人数字段
	DBModel
}

func (c *ScheduleCar) getNoSeatCount(seatBit uint64, depIdx, arrIdx uint8) uint8 {
	// 车厢未设置站票时，直接返回 0
	if c.NoSeatCount == 0 {
		return 0
	}
	// 非站票数与站票数之和
	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	c.RUnlock()
	return uint8(totalSeatCount) - maxTravelerCountInRoute
}

// getAvailableSeat 获取可预订的座位,是否获取成功标记,是否为拼凑的站票标记
func (c *ScheduleCar) getAvailableSeat(depIdx, arrIdx uint8, isStudent bool) (s *ScheduleSeat, ok bool) {
	seatBit := countSeatBit(depIdx, arrIdx)
	for i := 0; i < len(c.Seats); i++ {
		if c.Seats[i].Book(seatBit, isStudent) {
			c.occupySeat(depIdx, arrIdx)
			return &c.Seats[i], true
		}
	}
	return nil, false
}

func (c *ScheduleCar) getAvailableNoSeat(depIdx, arrIdx uint8) (s *ScheduleSeat, ok bool) {
	if c.NoSeatCount == 0 {
		return nil, false
	}
	seatBit := countSeatBit(depIdx, arrIdx)
	// 下面开始查找拼凑的站票
	// 非站票数与站票数之和
	totalSeatCount := len(c.Seats) + int(c.NoSeatCount)
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	c.RUnlock()
	if totalSeatCount-int(maxTravelerCountInRoute) > 0 {
		c.occupySeat(depIdx, arrIdx)
		s = &ScheduleSeat{SeatBit: seatBit}
		return s, true
	}
	return nil, false
}

// 某座位被占用
func (c *ScheduleCar) occupySeat(depIdx, arrIdx uint8) {
	if c.NoSeatCount == 0 {
		return
	}
	var maxCount uint8
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxCount {
			maxCount = c.EachRouteTravelerCount[i]
		}
	}
	if maxCount >= uint8(len(c.Seats))+c.NoSeatCount {
		return
	}
	c.Lock()
	for i := depIdx; i < arrIdx; i++ {
		c.EachRouteTravelerCount[i]++
	}
	c.Unlock()
}

// 某座位被释放
func (c *ScheduleCar) releaseSeat(depIdx, arrIdx uint8) {
	if c.NoSeatCount == 0 {
		return
	}
	c.Lock()
	for i := depIdx; i < arrIdx; i++ {
		c.EachRouteTravelerCount[i]--
	}
	c.Unlock()
}

// ScheduleSeat 排班中的座位
type ScheduleSeat struct {
	SeatBit    uint64 // 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	sync.Mutex        // 锁，订票与退票均需要锁
	Seat
}

// IsAvailable 根据路段和乘客类型判断能否订票
func (s *ScheduleSeat) IsAvailable(seatBit uint64, isStudent bool) bool {
	// 学生可订成人票，成人不可订学生票，发车前，需将未售的学生票全部改为成人票，用以出售
	if s.IsStudent && !isStudent {
		return false
	}
	return s.SeatBit^seatBit == s.SeatBit+seatBit
}

// Book 订票
func (s *ScheduleSeat) Book(seatBit uint64, isStudent bool) bool {
	if !s.IsAvailable(seatBit, isStudent) {
		return false
	}
	s.Lock()
	s.SeatBit ^= seatBit
	s.Unlock()
	return true
}

// Release 退票或取消订单，释放座位对应路段的资源
func (s *ScheduleSeat) Release(seatBit uint64) {
	s.Lock()
	s.SeatBit ^= (^seatBit)
	s.Unlock()
}
