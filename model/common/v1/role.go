package model

// Role Type
type RoleType int32

const (
	RoleType_Admin      RoleType = 0
	RoleType_Dispensary RoleType = 1
	RoleType_Cultivator RoleType = 2
	RoleType_Processor  RoleType = 3
	RoleType_Auditor    RoleType = 4
)

// Enum value maps for RoleType.
var (
	RoleType_name = map[int32]string{
		0: "admin",
		1: "dispensary",
		2: "cultivator",
		3: "processor",
		4: "auditor",
	}
	RoleType_value = map[string]int32{
		"admin":      0,
		"dispensary": 1,
		"cultivator": 2,
		"processor":  3,
		"auditor":    4,
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
