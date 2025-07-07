package model

// Role Type
type RoleType int32

const (
	RoleType_Admin    RoleType = 0
	RoleType_Manager  RoleType = 1
	RoleType_User     RoleType = 2
	RoleType_Viewer   RoleType = 3
)

// Enum value maps for RoleType.
var (
	RoleType_name = map[int32]string{
		0: "admin",
		1: "manager",
		2: "user",
		3: "viewer",
	}
	RoleType_value = map[string]int32{
		"admin":   0,
		"manager": 1,
		"user":    2,
		"viewer":  3,
	}
)

type Role struct {
	Id       int32    `gorm:"primaryKey;autoIncrement:true;column:id" json:"id"`
	Original bool     `gorm:"column:original" json:"original"` // Default system Roles (can't be changed) or deleted
	RoleType RoleType `gorm:"column:role_type" json:"role_type"`
	Desc     string   `gorm:"column:desc;unique" json:"desc"`
	Status   string   `gorm:"column:status" json:"status"`
	CommonModel
}
