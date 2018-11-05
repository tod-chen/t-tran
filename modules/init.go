package modules

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	// mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"gopkg.in/mgo.v2"
)

const (
	constMgoDB = "t-tran"
)

var (
	mgoSession *mgo.Session
	db         *gorm.DB
)

// Config 配置项
type Config struct {
	Key   string // 配置名
	Value string // 配置值
}

// DBModel 数据库模型
type DBModel struct {
	// 主键
	ID       uint64    `gorm:"primary_key;AUTO_INCREMENT(1)" json:"id"`
	CreateAt time.Time `gorm:"type:datetime"` // 创建时间
	UpdateAt time.Time `gorm:"type:datetime"` // 更新时间
}

func getMgoSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial("localhost:27017")
		if err != nil {
			panic(err)
		}
		mgoSession.SetMode(mgo.Monotonic, true)
	}
	return mgoSession.Clone()
}

func init() {
	fmt.Println("init modules beginning")
	defer fmt.Println("init modules end")
	var err error
	db, err = gorm.Open("mysql", "root:@/t-tran?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

	if !db.HasTable(&Config{}) {
		fmt.Println("create tables beginning")
		defer fmt.Println("create tables end")
		db.CreateTable(&Config{})
		db.CreateTable(&Station{})
		db.CreateTable(&TranInfo{})
		db.CreateTable(&Route{})
		db.CreateTable(&RoutePrice{})
		db.CreateTable(&Car{})
		db.CreateTable(&Seat{})
		// db.CreateTable(&ScheduleTran{})
		// db.CreateTable(&ScheduleCar{})
		// db.CreateTable(&ScheduleSeat{})
		db.CreateTable(&Order{})
		db.CreateTable(&User{})
		db.CreateTable(&Contact{})
	}

	sTran := &ScheduleTran{TranNum: "C2220"}
	fmt.Println(sTran)

	initGoPool()
	initStation()
	initTranInfo()
	initTranSchedule()

	fmt.Println("init done...")
}
