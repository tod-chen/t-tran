package modules

import (
	"errors"
)

// Passenger 乘客
type Passenger struct {
	Name           string // 姓名
	IsMale         bool   // 性别
	Area           string // 国家地区
	PaperworkType  uint8  // 证件类型
	PaperworkNum   string // 证件号码
	PaperworkValid bool   // 证件是否有效
	PassengerType  uint8  // 乘客类型
	PhoneNum       string // 手机号
	TelNum         string // 固话
	Email          string // 邮箱
	Addr           string // 地址
	ZipCode        string // 邮编
	DBModel
}

// User 注册用户
type User struct {
	UserName string `gorm:"type:nvarchar(50)"` //用户名
	Password string `gorm:"type:varchar(50)"`  //密码
	PhoneNumValid bool      // 手机号是否有效
	EmailValid    bool      // 邮箱是否有效
	Contacts      []Contact `gorm:"foreignkey:UserID"` // 联系人
	Passenger
}

// Register 注册
func (u *User) Register() error {
	count := 0
	db.Model(&User{}).Where("(user_name = ? or (paperwork_num = ? and paperwork_type = ?)) and paperwork_valid = 1", u.UserName, u.PaperworkNum, u.PaperworkType).Count(&count)
	if count != 0 {
		return errors.New("用户名或证件信息已存在")
	}
	db.Create(u)
	return nil
}

// ChangePwd 修改密码
func (u *User) ChangePwd(newPwd string) error {
	db.Model(u).Update("password", newPwd)
	return nil
}

// Edit 修改个人信息
func (u *User) Edit() error {
	db.Save(u)
	return nil
}

// Contact 常用联系人
type Contact struct {
	UserID uint `gorm:"index:main"` // 用户ID
	Passenger
}

// Add 添加联系人
func (c *Contact) Add() error {
	db.Create(c)
	return nil
}

// Edit 修改联系人信息
func (c *Contact) Edit() error {
	db.Save(c)
	return nil
}

// Remove 删除联系人
func (c *Contact) Remove() error {
	db.Delete(c)
	return nil
}
