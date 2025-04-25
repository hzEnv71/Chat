package models

import (
	"fmt"
	"ginchat/utils"
	"gorm.io/gorm"
	"time"
)

type UserBasic struct {
	gorm.Model
	Name          string
	PassWord      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Avatar        string //头像
	Identity      string //用于验证密码
	ClientIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool
	DeviceInfo    string
}

func (user *UserBasic) TableName() string { return "user_basic" }

// GetUserList 查询所有用户
func GetUserList() []*UserBasic {
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	return data
}

// FindUserByNameAndPwd 查询用户通过姓名和密码
func FindUserByNameAndPwd(name string, password string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ? and pass_word=?", name, password).First(&user)
	//token加密
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.MD5Encode(str)
	utils.DB.Model(&user).Where("id = ?", user.ID).Update("identity", temp)
	return user
}

// FindUserByName 查询用户通过姓名
func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ?", name).First(&user)
	return user
}

// FindUserByID FindByID 查找用户通过ID
func FindUserByID(id uint) UserBasic {
	user := UserBasic{}
	utils.DB.Where("id = ?", id).First(&user)
	return user
}

// FindUserByPhone 查询用户通过手机号
func FindUserByPhone(phone string) *gorm.DB {
	user := UserBasic{}
	return utils.DB.Where("Phone = ?", phone).First(&user)
}

// FindUserByEmail 查询用户通过邮箱
func FindUserByEmail(email string) *gorm.DB {
	user := UserBasic{}
	return utils.DB.Where("email = ?", email).First(&user)
}

// CreateUser 创建用户
func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}

// DeleteUser 删除用户
func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}

// UpdateUser 更新用户
func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, PassWord: user.PassWord, Phone: user.Phone, Email: user.Email, Avatar: user.Avatar})
}
