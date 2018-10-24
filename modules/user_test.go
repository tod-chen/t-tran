package modules

import (
	"fmt"
	"testing"
)

func TestStructUser(t *testing.T) {
	u := &User{
		UserName:   "tod-chen",
		Password:   "pwd",
		EmailValid: true,
		Passenger: Passenger{
			Name:           "陈德亭",
			IsMale:         true,
			Area:           "China",
			PaperworkType:  1,
			PaperworkNum:   "420xxxxxxxNNNNmmmm",
			PaperworkValid: true,
			PassengerType:  1,
			PhoneNum:       "189xxxxxxxx",
			TelNum:         "",
			Email:          "tod-chen@foxmail.com",
			Addr:           "",
			ZipCode:        "",
		},
	}
	
	err := u.Register()
	if err == nil{
		t.Log("Register pass for create")
	} else if err.Error() == "用户名或证件信息已存在"  {
		t.Log("Register pass for exist")
	} else {
		t.Error("Register fail")
	}

	if err := u.ChangePwd("new pwd"); err == nil {
		var tempUser User
		db.Where("user_name = ? and paperwork_num = ? and paperwork_type = ?", u.UserName, u.PaperworkNum, u.PaperworkType).First(&tempUser)
		fmt.Print(tempUser)
		if tempUser.Password == "new pwd" {
			t.Log("ChangePwd pass")
		} else {
			t.Error("ChangePwd fail for db value")
		}
	} else {
		t.Error("ChangePwd fail for option")
	}
}
