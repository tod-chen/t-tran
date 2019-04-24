package modules

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	// 排班的车厢集
	scheduleCarMap map[int](*ScheduleCar)
	// 列车按日期安排表
	scheduleTranMap sync.Map
	// 排班缓存存放处
	scheduleCache scheduleTranCache
)

type scheduleTranCache struct {
	mod   int
	cap   int
	cache [][]*ScheduleTran
	sync.Once
	sync.RWMutex
	ticker <-chan time.Time
}

func (s *scheduleTranCache) init(length, eleCap int) {
	s.Do(func() {
		s.mod = length
		s.cap = eleCap
		s.cache = make([]([]*ScheduleTran), s.mod)
		for i := 0; i < s.mod; i++ {
			s.cache[i] = make([]*ScheduleTran, 0, s.cap)
		}
		fmt.Println("scheduleTranCache init done")
		s.ticker = time.Tick(time.Second)
		for now := range s.ticker {
			trans := s.cache[now.Second()%s.mod]
			session := getMgoSession()
			coll := session.DB(constMgoDB).C("tranSchedule")
			for i := 0; i < len(trans); i++ {
				if trans[i].hasChanged {
					trans[i].LastUpdateTime = now
					coll.Update(bson.M{"tranNum": trans[i].TranNum, "departureDate": trans[i].DepartureDate}, &trans[i])
				} else if trans[i].LastUpdateTime.Sub(time.Now()) > 10*time.Minute {
					// 10分钟内无变更，缓存移除
					scheduleTranMap.Delete(trans[i].TranNum + "_" + trans[i].DepartureDate)
					trans = append(trans[:i], trans[i+1:]...)
				}
			}
			s.cache[now.Second()%s.mod] = trans
			session.Close()
		}
	})
}

func (s *scheduleTranCache) getScheduleTran(tranNum, date string) *ScheduleTran {
	key := tranNum + "_" + date
	if val, ok := scheduleTranMap.Load(key); ok {
		return val.(*ScheduleTran)
	}
	tran := &ScheduleTran{}
	session := getMgoSession()
	coll := session.DB(constMgoDB).C("tranSchedule")
	err := coll.Find(bson.M{"tranNum": tranNum, "departureDate": date}).One(tran)
	session.Close()
	if err == nil {
		tn, _ := strconv.Atoi(tranNum[1 : len(tranNum)-1])
		sli := s.cache[tn%s.mod]
		s.cache[tn%s.mod] = append(sli, tran)
		scheduleTranMap.Store(key, tran)
	}
	return tran
}

func initSchedule() {
	initScheduleCar()
	initScheduleTran()
	scheduleCache = scheduleTranCache{}
	go scheduleCache.init(30, 5)
}

// 初始化排班的车厢
func initScheduleCar() {
	start := time.Now()
	scheduleCarMap = make(map[int](*ScheduleCar), len(carMap))
	for id, car := range carMap {
		sc := ScheduleCar{
			SeatType:    car.SeatType,
			NoSeatCount: car.NoSeatCount,
			Seats:       make([]ScheduleSeat, len(car.Seats)),
		}
		for i := 0; i < len(car.Seats); i++ {
			sc.Seats[i].SeatNum = car.Seats[i].SeatNum
			sc.Seats[i].IsStudent = car.Seats[i].IsStudent
		}
		scheduleCarMap[id] = &sc
	}
	fmt.Println("init schedule car complete, cost time:", time.Now().Sub(start).Seconds(), "(s)")
}

// 初始化列车排班
func initScheduleTran() {
	start := time.Now()
	goPoolCompute := newGoPool(300)
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	for idx := 0; idx < len(tranInfos); idx++ {
		if !tranInfos[idx].IsSaleTicket {
			continue
		}
		wg.Add(1)
		goPoolCompute.Take()
		go func(i int) {
			session := getMgoSession()
			coll := session.DB(constMgoDB).C("tranSchedule")
			defer func() {
				session.Close()
				goPoolCompute.Return()
				wg.Done()
			}()
			// 排班的开始日期和截止日期
			start, end := today, today.AddDate(0, 0, constDays)
			if tranInfos[i].EnableStartDate.After(start) {
				start = tranInfos[i].EnableStartDate
			} else if tranInfos[i].ScheduleDays != 1 {
				// 当前时间晚于配置生效的起始时间，且并不是每天发车，则将排班开始日期设置为下次排班日期
				start = tranInfos[i].EnableStartDate
				for start.Before(today) {
					start = start.AddDate(0, 0, tranInfos[i].ScheduleDays)
				}
			}
			if tranInfos[i].EnableEndDate.Before(end) {
				end = tranInfos[i].EnableEndDate
			}
			sTran := ScheduleTran{}
			// 找到当前车次，已初始化的最后一个排班，并以该排班的发车日期作为开始时间
			if err := coll.Find(bson.M{"tranNum": tranInfos[i].TranNum}).Sort("-departureDate").One(&sTran); err == nil {
				lastDepDate, _ := time.Parse(ConstYmdFormat, sTran.DepartureDate)
				if lastDepDate.After(tranInfos[i].EnableEndDate) {
					// 当前车次最后一个排班的日期，晚于配置的截止日期，说明该配置的排班已经排好了，无需处理
					return
				}
				if start.Before(lastDepDate) {
					// 当前车次配置的排班未排完，则以下次排班日期作为开始日期
					start = lastDepDate.AddDate(0, 0, tranInfos[i].ScheduleDays)
				}
			} else { // 当前车次无排班，。设置好其他信息
				sTran = ScheduleTran{
					TranNum:     tranInfos[i].TranNum,
					Cars:        tranInfos[i].getScheduleCars(),
					FullSeatBit: countSeatBit(0, uint8(len(tranInfos[i].Timetable)-1)),
				}
			}
			for day := start; !day.After(end); day = day.AddDate(0, 0, tranInfos[i].ScheduleDays) {
				y, M, d := day.Date()
				h, m, s := tranInfos[i].SaleTicketTime.Clock()
				sTran.DepartureDate = day.Format(ConstYmdFormat)
				sTran.SaleTicketTime = time.Date(y, M, d-constDays, h, m, s, 0, time.Local)
				sTran.LastUpdateTime = time.Now()
				if err := coll.Insert(&sTran); err != nil {
					panic(err)
				}
			}
		}(idx)
	}
	wg.Wait()
	goPoolCompute.Close()
	fmt.Println("init scheduleTran complete, cost time:", time.Now().Sub(start).Seconds(), "(s)")
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

func buildResidualTicketInfo(t *TranInfo, depIdx, arrIdx uint8, date string, isStudent bool) *ResidualTicketInfo {
	r := &ResidualTicketInfo{
		tranNum:  t.TranNum,
		date:     date,
		depIdx:   depIdx,
		depCode:  t.Timetable[depIdx].StationCode,
		depTime:  t.Timetable[depIdx].DepTime,
		arrIdx:   arrIdx,
		arrCode:  t.Timetable[arrIdx].StationCode,
		arrTime:  t.Timetable[arrIdx].ArrTime.Format(ConstHmFormat),
		costTime: t.Timetable[arrIdx].ArrTime.Sub(t.Timetable[depIdx].DepTime).String(),
	}
	if t.IsSaleTicket { // 车次配置中，售票标记为真
		st := scheduleCache.getScheduleTran(t.TranNum, r.date)
		if st.SaleTicketTime.Before(time.Now()) { // 已过售票时间，计算各席位的余票数
			r.seatCount = st.GetAvaliableSeatCount(t, r.depIdx, r.arrIdx, isStudent)
		} else { // 未到售票时间，调整备注
			r.remark = st.SaleTicketTime.Format(ConstYMdHmFormat)
		}
	} else {
		r.remark = t.NonSaleRemark
	}
	return r
}

// 结果转为字符串
func (r *ResidualTicketInfo) toString() string {
	list := []string{r.tranNum, r.date,
		strconv.Itoa(int(r.depIdx)), r.depCode, r.depTime.Format(ConstHmFormat),
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
	DepartureDate  string        `bson:"departureDate"`  // 发车日期
	TranNum        string        `bson:"tranNum"`        // 车次号
	SaleTicketTime time.Time     `bson:"saleTicketTime"` // 售票时间
	Cars           []ScheduleCar `bson:"cars"`           // 车厢
	FullSeatBit    int64         `bson:"fullSeatBit"`    // 全程满座的位标记值，某座位的位标记与此值相等时，表示该座位全程满座了
	hasChanged     bool          // 缓存是否有变更
	LastUpdateTime time.Time     `bson:"lastUpdateTime"` // 最后更新时间
}

// Save 保存到数据库
func (st *ScheduleTran) Save() (bool, string) {
	st.SaleTicketTime = st.SaleTicketTime.Add(-8 * time.Hour)
	db.Save(st)
	return true, ""
}

// GetAvaliableSeatCount 获取各席别余票数
func (st *ScheduleTran) GetAvaliableSeatCount(t *TranInfo, depIdx, arrIdx uint8, isStudent bool) []int {
	// 总共11类座次
	result := []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	// 路段位标记
	seatBit, noSeatCount := countSeatBit(depIdx, arrIdx), 0
	for seatType, idxs := range t.carTypeIdxMap {
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
		result[seatTypeIdxMap[seatType]] = avaliableSeatCount
	}
	result[len(result)-1] = noSeatCount
	return result
}

// ScheduleCar 排班中的车厢
type ScheduleCar struct {
	SeatType               string         // 座次
	CarNum                 uint8          // 车厢号
	NoSeatCount            uint8          // 车厢内站票数
	Seats                  []ScheduleSeat // 车厢的所有座位
	EachRouteTravelerCount []uint8        // 各路段乘客人数，用于计算可拼凑的站票数，仅在有站票的车厢使用
	sync.RWMutex                          // 读写锁，用于保护各路段乘客人数字段
}

func (c *ScheduleCar) getNoSeatCount(depIdx, arrIdx uint8) uint8 {
	// 车厢未设置站票时，直接返回 0
	if c.NoSeatCount == 0 {
		return 0
	}
	// 非站票数与站票数之和
	totalSeatCount := uint8(len(c.Seats)) + c.NoSeatCount
	c.RLock()
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	for i := depIdx; i < arrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	c.RUnlock()
	return totalSeatCount - maxTravelerCountInRoute
}

// getAvailableSeat 获取可预订的座位,是否获取成功标记,是否为拼凑的站票标记
func (c *ScheduleCar) getAvailableSeat(par *SubmitOrderModel) (s *ScheduleSeat, seatIdx uint8, ok bool) {
	for i := 0; i < len(c.Seats); i++ {
		if c.Seats[i].Book(par.seatBit, par.IsStudent) {
			c.occupySeat(par.DepIdx, par.ArrIdx)
			return &c.Seats[i], uint8(i), true
		}
	}
	return nil, 0, false
}

func (c *ScheduleCar) getAvailableNoSeat(par *SubmitOrderModel) (s *ScheduleSeat, ok bool) {
	if c.NoSeatCount == 0 {
		return nil, false
	}
	// 下面开始查找拼凑的站票
	// 非站票数与站票数之和
	totalSeatCount := uint8(len(c.Seats)) + c.NoSeatCount
	// 旅途中当前车厢旅客最大数
	var maxTravelerCountInRoute uint8
	c.RLock()
	for i := par.DepIdx; i < par.ArrIdx; i++ {
		if c.EachRouteTravelerCount[i] > maxTravelerCountInRoute {
			maxTravelerCountInRoute = c.EachRouteTravelerCount[i]
		}
	}
	c.RUnlock()
	if totalSeatCount-maxTravelerCountInRoute > 0 {
		c.occupySeat(par.DepIdx, par.ArrIdx)
		s = &ScheduleSeat{SeatBit: par.seatBit}
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
	SeatNum   string // 座位号
	IsStudent bool   // 是否预留给学生的座位
	SeatBit   int64  // 座位的位标记，64位代表64个路段，值为7时，表示从起始站到第四站，这个座位都被人订了
}

// IsAvailable 根据路段和乘客类型判断能否订票
func (s *ScheduleSeat) IsAvailable(seatBit int64, isStudent bool) bool {
	// 学生可订成人票，成人不可订学生票，发车前，需将未售的学生票全部改为成人票，用以出售
	if s.IsStudent && !isStudent {
		return false
	}
	return s.SeatBit^seatBit == s.SeatBit+seatBit
}

// Book 订票
func (s *ScheduleSeat) Book(seatBit int64, isStudent bool) bool {
	if !s.IsAvailable(seatBit, isStudent) {
		return false
	}
	return atomic.CompareAndSwapInt64(&s.SeatBit, s.SeatBit, s.SeatBit+seatBit)
}

// Release 退票或取消订单，释放座位对应路段的资源
func (s *ScheduleSeat) Release(seatBit int64) {
	atomic.AddInt64(&s.SeatBit, -seatBit)
}
