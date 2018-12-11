package modules

import (
	"errors"
	"time"
)

// Passenger 乘客
type Passenger struct {
	ID            uint64
	Name          string // 姓名
	Gender        bool   // 性别
	Area          string // 国家地区
	PaperworkType uint8  // 证件类型
	PaperworkNum  string // 证件号码
	Status        uint8  // 乘客信息状态 0:待核验; 1:核验通过; 2:核验未通过; 3:黑名单; ...etc
	PassengerType uint8  // 乘客类型
	PhoneNum      string // 手机号
	TelNum        string // 固话
	Email         string // 邮箱
	Addr          string // 地址
	ZipCode       string // 邮编
}

// PassengerAdderMap 乘客添加者的映射关系表
type PassengerAdderMap struct {
	PID uint64 // 乘客ID
	UID uint64 // 添加者ID
}

// TODO: Passenger 状态变更后，要通知所有添加者，同时变更对应的状态
// func (p *Passager) StatusChanged(newStatus uint8) error

// Approve 乘客信息核验通过
func (p *Passenger) Approve() error {
	p.Status = 1

	return nil
}

// Disapprove 乘客信息核验未通过
func (p *Passenger) Disapprove() error {
	p.Status = 2
	return nil
}

// User 注册用户
type User struct {
	UserName  string    `gorm:"type:nvarchar(50)"` //用户名
	Password  string    `gorm:"type:varchar(50)"`  //密码
	LoginTime time.Time `gorm:"type:datetime"`     // 登录时间，可用于判断session超时
	Passenger           // 注册用户也是乘客
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
	UserID    uint      `gorm:"index:main"` // 用户ID
	Passenger           // 联系人是乘客
	AddDate   time.Time // 添加的日期
}

// QueryContact 查询联系人
// uid 用户ID
// name 联系人姓名，模糊查询
func QueryContact(uid uint64, name string) (list []Contact) {
	db.Where("user_id = ? and name like ?", uid, "%"+name+"%").Find(&list)
	return
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
