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
	// 座次类型及前端显示序号关系
	seatTypeIdxMap = map[string](int){
		constSeatTypeSpecial:             0,
		constSeatTypeFristClass:          1,
		constSeatTypeSecondClass:         2,
		constSeatTypeAdvancedSoftSleeper: 3,
		constSeatTypeSoftSleeper:         4,
		constSeatTypeEMUSleeper:          5,
		constSeatTypeMoveSleeper:         6,
		constSeatTypeHardSleeper:         7,
		constSeatTypeSoftSeat:            8,
		constSeatTypeHardSeat:            9,
	}
	// 列车按日期安排表
	scheduleTranMap map[string]([]ScheduleTran)
)

func isSameDay(src, tar time.Time) bool {
	sy, sm, sd := src.Date()
	ty, tm, td := tar.Date()
	return sy == ty && sm == tm && sd == td
}

// 初始化列车排班
func initTranSchedule() {
	scheduleTranMap = make(map[string]([]ScheduleTran), constDays)
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	session := getMgoSession()
	defer session.Close()
	for i := 0; i < len(tranInfos); i++ {
		fmt.Println("init schedule tran count: ", i, ", tranNum:", tranInfos[i].TranNum)
		if !tranInfos[i].IsSaleTicket {
			continue
		}
		start := now
		// 非每天发车的车次
		if tranInfos[i].ScheduleDays != 1 {
			start = tranInfos[i].EnableStartDate
			// 确定该车次在今后一个月内的首次发车时间
			for d := 0; ; d += tranInfos[i].ScheduleDays {
				tempDay := start.AddDate(0, 0, d)
				if !tempDay.Before(today) {
					start = tempDay
					break
				}
			}
		}
		// 路段出发间隔天数不为零，则初次发车时间应向前推，使该车次的最后一个路段在今后一个月的时间段内
		if tranInfos[i].RouteDepDurationDays != 0 {
			durDay := tranInfos[i].RouteDepDurationDays
			lastRouteDepDay := start.AddDate(0, 0, durDay)
			for d := tranInfos[i].ScheduleDays; ; d += tranInfos[i].ScheduleDays {
				tempDay := lastRouteDepDay.AddDate(0, 0, -d)
				if tempDay.Before(today) {
					break
				}
				if tempDay.AddDate(0, 0, -durDay).After(tranInfos[i].EnableStartDate) {
					start = tempDay.AddDate(0, 0, -durDay)
				}
			}
		}
		end := now.AddDate(0, 0, constDays)
		if tranInfos[i].EnableEndDate.Before(end) {
			end = tranInfos[i].EnableEndDate
		}
		for day := start; day.Before(end) || isSameDay(day, end); day = day.AddDate(0, 0, tranInfos[i].ScheduleDays) {
			initScheduleByDate(day, &tranInfos[i])
		}
	}
	fmt.Println("tran schedule count:", len(scheduleTranMap))
	fmt.Println(scheduleTranMap["2018-11-07"][1])
	fmt.Println("init scheduleTran complete")
}

func initScheduleByDate(date time.Time, t *TranInfo) {
	strDate, sTran := date.Format(constYmdFormat), &ScheduleTran{}
	tranScheduleColl := getMgoSession().DB(constMgoDB).C("tranSchedule")
	fmt.Println(strDate)
	q := tranScheduleColl.Find(bson.M{"departureDate": strDate, "tranNum": t.TranNum})
	err := q.One(sTran)
	// 未找到，则新增一个
	if err != nil {
		y, M, d := date.Date()
		h, m, s := t.SaleTicketTime.Clock()
		sTran = &ScheduleTran{
			DepartureDate:  strDate,
			TranNum:        t.TranNum,
			SaleTicketTime: time.Date(y, M, d, h, m, s, 0, time.Local),
			Cars:           *t.getScheduleCars(),
			FullSeatBit:    countSeatBit(0, uint8(len(t.Timetable)-1)),
		}
		sTran.carTypeIdxMap = make(map[string]([]uint8))
		for i := 0; i < len(sTran.Cars); i++ {
			idxs, exist := sTran.carTypeIdxMap[sTran.Cars[i].SeatType]
			if !exist {
				idxs = make([]uint8, 0)
			}
			idxs = append(idxs, uint8(i))
			sTran.carTypeIdxMap[sTran.Cars[i].SeatType] = idxs
		}
		if err = tranScheduleColl.Insert(&sTran); err != nil {
			panic(err)
		}
	}
	trans := scheduleTranMap[strDate]
	trans = append(trans, *sTran)
	scheduleTranMap[strDate] = trans
}

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
	r.seatCount = *st.GetSeatCount(r.depIdx, r.arrIdx, isStudent)
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
	DepartureDate  string               `bson:"departureDate"`  // 发车日期
	TranNum        string               `bson:"tranNum"`        // 车次号
	SaleTicketTime time.Time            `bson:"saleTicketTime"` // 售票时间
	Cars           []ScheduleCar        `bson:"cars"`           // 车厢
	carTypeIdxMap  map[string]([]uint8) `bson:"carTypeIdxMap"`  // 各座次类型及其对应的车厢索引集合
	FullSeatBit    uint64               `bson:"fullSeatBit"`    // 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
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
func (st *ScheduleTran) GetSeatCount(depIdx, arrIdx uint8, isStudent bool) *[]int {
	// 总共11类座次
	result := []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	// 路段位标记
	seatBit, noSeatCount := countSeatBit(depIdx, arrIdx), 0
	for seatType, idxs := range st.carTypeIdxMap {
		resultIdx := seatTypeIdxMap[seatType]
		avaliableSeatCount := 0
		for _, idx := range idxs {
			for i := 0; i < len(st.Cars[idx].Seats) && avaliableSeatCount < constMaxAvaliableSeatCount; i++ {
				if st.Cars[idx].Seats[i].IsAvailable(seatBit, isStudent) {
					avaliableSeatCount++
				}
			}
			if noSeatCount < constMaxAvaliableSeatCount {
				noSeatCount += int(st.Cars[idx].getNoSeatCount(depIdx, arrIdx))
			}
		}
		result[resultIdx] = avaliableSeatCount
	}
	result[len(result)-1] = noSeatCount
	return &result
}

// ScheduleCar 排班中的车厢
type ScheduleCar struct {
	SeatType    string         // 座次
	CarNum      uint8          // 车厢号
	NoSeatCount uint8          // 车厢内站票数
	Seats       []ScheduleSeat // 车厢的所有座位
	//EachRouteTravelerCountStr string         // 各路段乘客人数，用英文逗号分隔存于数据库
	// 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	EachRouteTravelerCount []uint8 //`gorm:"-"`
	sync.RWMutex                   // 读写锁，用于保护各路段乘客人数字段
	DBModel
}

func (c *ScheduleCar) getNoSeatCount(depIdx, arrIdx uint8) uint8 {
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
	SeatNum    string
	IsStudent  bool
	SeatBit    uint64     // 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
	sync.Mutex `bson:"-"` // 锁，订票与退票均需要锁
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
