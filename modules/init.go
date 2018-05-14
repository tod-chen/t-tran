package modules

import (
	"database/sql"
	// mysql
	_ "github.com/go-sql-driver/mysql"

	"gopkg.in/mgo.v2"
)

var (
	mgoSession *mgo.Session
	mgoDbName  = "t-tran"
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
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/t-tran?charset=utf8")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	initStation(db)
	initTran()
	initGoPool()
}
