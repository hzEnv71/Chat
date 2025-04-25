package models

import (
	"fmt"
	"ginchat/utils"
	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	Name    string
	OwnerId uint //属于哪个用户  用户id(群主)
	Img     string
	Desc    string
}

func (community *Community) TableName() string {
	return "communities"
}

// CreateCommunity 创建群聊
func CreateCommunity(community Community) (int, string) {
	tx := utils.DB.Begin() //事务一旦开始，不论什么异常最终都会 Rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println("群聊创建失败:", err)
		tx.Rollback()
		return -1, "建群失败"
	}
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2 //群关系
	if err := utils.DB.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "创建群关系失败"
	}
	tx.Commit()
	return 0, "建群成功"
}

// LoadCommunity 加载群聊
func LoadCommunity(ownerId uint) ([]*Community, string) {
	contacts := make([]Contact, 0)
	communityIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=2", ownerId).Find(&contacts) //type=2 群关系
	for _, v := range contacts {
		communityIds = append(communityIds, uint64(v.TargetId))
	}
	data := make([]*Community, 10)
	utils.DB.Where("id in ?", communityIds).Find(&data)
	return data, "查询成功"
}

// JoinGroup 加入群聊
func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2
	community := Community{}

	utils.DB.Where("id=? or name=?", comId, comId).Find(&community) //群名称或者群号(id)
	if len(community.Name) == 0 {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, comId).Find(&contact)
	if contact.CreatedAt.IsZero() {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	} else {
		return -1, "已加过此群"
	}
}
