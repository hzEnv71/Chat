package models

import "gorm.io/gorm"

// GroupBasic 群信息
type GroupBasic struct {
	gorm.Model
	Name    string
	OwnerId uint
	Icon    string
	Type    int
	Desc    string
}

func (groupBasic *GroupBasic) TableName() string {
	return "group_basic"
}
