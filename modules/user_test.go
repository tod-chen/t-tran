package modules

import (
	"testing"
)

func TestStructPassenger(t *testing.T){
	p, ok := getPassenger("420116199405293568", 1) 
	if ok && p.PaperworkNum == "420116199405293568" {
		t.Log("get passenger success")
	} else {
		t.Log("get passenger fail")
	}

	
}

func TestStructUser(t *testing.T) {
	u := &User{
		UserName: "tod-chen",
		Password: "123456",
	}
	p := &Passenger{
		Name:          "陈德亭",
		IsMale:        true,
		Area:          "China",
		PaperworkType: 1,
		PaperworkNum:  "420xxxxxxxNNNNmmmm",
		PassengerType: 1,
		PhoneNum:      "189xxxxxxxx",
		TelNum:        "",
		Email:         "tod-chen@foxmail.com",
		Addr:          "",
		ZipCode:       "",
	}

	success, err := u.Register(*p)
	if success {
		t.Log("Register pass for create")
	} else if err.Error() == "用户名或证件信息已存在" {
		t.Log("Register pass for exist")
	} else {
		t.Error("Register fail")
	}

	if err := u.ChangePwd("new pwd"); err == nil {
		var tempUser User
		db.Where("user_name = ?", u.UserName).First(&tempUser)
		if tempUser.Password == "new pwd" {
			t.Log("ChangePwd pass")
		} else {
			t.Error("ChangePwd fail for db value")
		}
	} else {
		t.Error("ChangePwd fail for option")
	}	
}



