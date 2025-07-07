package model

type InboxStatus int32

const (
	InboxStatus_normal InboxStatus = 0
	InboxStatus_demo   InboxStatus = 1
)

// Enum value maps for InboxStatus.
var (
	InboxStatus_name = map[int32]string{
		0: "received",
		1: "open",
		2: "readed",
	}
	InboxStatus_value = map[string]int32{
		"normal": 0,
		"demo":   1,
		"readed": 2,
	}
)

type Inbox struct {
	Id         string      `gorm:"primaryKey;autoIncrement:true;column:id" json:"id"`
	UserId     int32       `gorm:"column:user_id" json:"user_id"`
	User       User        `gorm:"foreignKey:UserId"`
	FromUserId int32       `gorm:"column:from_user_id" json:"from_user_id"`
	Subject    string      `gorm:"column:subject" json:"subject"`
	Body       string      `gorm:"column:body" json:"body"`
	Status     InboxStatus `gorm:"column:status" json:"status"`
	CommonModel
}
