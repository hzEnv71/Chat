package models

import (
	"ginchat/utils"

	"gorm.io/gorm"
)

// Contact 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息 用户id
	TargetId uint //用户属于哪个群 群id
	Type     int  //对应的类型  1好友  2群  3xx
	Desc     string
}

func (contact *Contact) TableName() string {
	return "contact"
}

// SearchFriend 查找好友通过id
func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	userIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		userIds = append(userIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", userIds).Find(&users)
	return users
}

// AddFriend 添加好友通过姓名
func AddFriend(userId uint, targetName string) (int, string) {
	if len(targetName) == 0 {
		return -1, "好友名称不能为空"
	}
	targetUser := FindUserByName(targetName)
	if targetUser.Salt != "" {
		if targetUser.ID == userId {
			return -1, "不能加自己"
		}
		contact := Contact{}
		utils.DB.Where("owner_id =?  and target_id =? and type=1", userId, targetUser.ID).Find(&contact)
		if contact.ID > 0 {
			return -1, "不能重复添加"
		}
		tx := utils.DB.Begin()
		defer func() { //事务一旦开始，不论什么异常最终都会 Rollback
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		var contact0, contact1 Contact
		contact0.OwnerId, contact0.TargetId, contact0.Type = userId, targetUser.ID, 1
		if err := utils.DB.Create(&contact0).Error; err != nil {
			tx.Rollback()
			return -1, "添加好友失败"
		}
		contact1.OwnerId, contact1.TargetId, contact1.Type = targetUser.ID, userId, 1
		if err := utils.DB.Create(&contact1).Error; err != nil {
			tx.Rollback()
			return -1, "添加好友失败"
		}
		tx.Commit()
		return 0, "添加好友成功"
	}
	return -1, "没有找到此用户"
}

// SearchUserByGroupId 查找所有用户通过群聊id
func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	userIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		userIds = append(userIds, v.OwnerId)
	}
	return userIds
}
