package modules

import (
	"fmt"

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
	defer func() {
		fmt.Println("init modules end")
		if p := recover(); p != nil {
			fmt.Printf("panic: %s\n", p)
		}
	}()
	var err error
	db, err = gorm.Open("mysql", "root:@/t-tran?charset=utf8&parseTime=True&loc=Asia%2FShanghai")
	if err != nil {
		panic(err)
	}
	initStation()
	initTranInfo()
	initSchedule()
	// 需要初始化用户数据，则取消下面一行代码的注释
	// initUserInfos()
}
