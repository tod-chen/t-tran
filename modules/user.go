package modules

import (
	"time"
)

type user struct{
	id int
	userName string
	password string
	phoneNum string
	emailAddr string
	contactIds []int
}

type contact struct{
	id int
	realName string
	paperworkType string
	paperworkNum string
	phoneNum string
	contactType string
	approveStatus int8
	addTime time.Time
}