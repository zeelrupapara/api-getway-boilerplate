package model

import (
	"fmt"

	"gorm.io/gorm"
)

type ResourceType int32

const (
	ResourceType_api    ResourceType = 0
	ResourceType_page   ResourceType = 1
	ResourceType_screen ResourceType = 2
)

var (
	ResourceType_name = map[ResourceType]string{
		0: "api",
		1: "page",
		2: "screen",
	}
	ResourceType_value = map[string]ResourceType{
		"api":    0,
		"page":   1,
		"screen": 2,
	}
)

type Action struct {
	Id         int32  `gorm:"primaryKey;autoIncrement:true;column:id" json:"id"`
	Desc       string `gorm:"column:desc" json:"desc"`
	ResourceId int32  `gorm:"column:resource_id" json:"resource_id"`
	Checked    bool   `gorm:"column:checked" json:"checked"`
}

type Resource struct {
	Id      int32        `gorm:"primaryKey;autoIncrement:true;column:id;" json:"id"`
	Type    ResourceType `gorm:"column:type" json:"type" validate:"required"`
	Desc    string       `gorm:"column:desc" json:"desc" validate:"required"`
	Status  string       `gorm:"column:status" json:"status"`
	Actions []Action     `gorm:"foreignKey:ResourceId;" json:"actions"`
}

// Group Model Validetor
func (r *Resource) ResourceValidetor(tx *gorm.DB) error {

	if _, ok := ResourceType_name[r.Type]; !ok {
		return fmt.Errorf("resouce type should be either an api, page or screen")
	}

	return nil
}

// gorm hooks runs everytime before creating new Group to validate
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	return r.ResourceValidetor(tx)
}

// gorm hooks runs everytime before updating the Group to validate
func (r *Resource) BeforeUpdate(tx *gorm.DB) error {
	return r.ResourceValidetor(tx)
}
