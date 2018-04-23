package modules

import (
	"database/sql"	
	_ "github.com/go-sql-driver/mysql"
)


func init(){	
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/t-tran?charset=utf8")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	initStation(db)
	initTran(db)
	initGoPool()
}


