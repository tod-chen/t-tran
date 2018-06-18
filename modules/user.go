package modules

import (
	"time"
)

// User 用户
type User struct {
	UserName   string        `gorm:"type:nvarchar(50)"` //用户名
	Password   string        `gorm:"type:varchar(50)"`  //密码
	PhoneNum   string        `gorm:"type:varchar(15)"`  //手机号
	EmailAddr  string        `gorm:"type:varchar(50)"`  //邮箱
	ContactIds string        `gorm:"type:varchar(500)"` // 联系人ID，用英文逗号分隔
	Contacts   []UserContact `gorm:"foreignkey:UserID"` // 联系人
	DBModel
}

// UserContact 常用联系人
type UserContact struct {
	UserID      uint      `gorm:"index:main"`       // 用户ID
	ContactType string    `gorm:"type:varchar(10)"` // 联系人类型：成人、儿童
	AddTime     time.Time `gorm:"type:datetime"`    // 联系人添加时间
	DBModel
}

// Contact 用户库
type Contact struct {
	RealName      string
	PaperworkType string `gorm:"index:q;type:nvarchar(10)"`
	PaperworkNum  string `gorm:"index:q;type:varchar(50)"`
	PhoneNum      string `gorm:"type:varchar(20)"`
	ApproveStatus uint8
	DBModel
}
