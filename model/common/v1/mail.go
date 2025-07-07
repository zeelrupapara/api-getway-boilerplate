package model

import (
	"time"

	"github.com/lithammer/shortuuid/v4"
	"gorm.io/gorm"
)

type MailStatus int32

const (
	MailStatus_queue   MailStatus = 0
	MailStatus_sent    MailStatus = 1
	MailStatus_pending MailStatus = 2
	MailStatus_draft   MailStatus = 3
	MailStatus_dropped MailStatus = 4
)

// Enum value maps for MailStatus.
var (
	MailStatus_name = map[int32]string{
		0: "queue",
		1: "sent",
		2: "pending",
		3: "draft",
		4: "dropped",
	}
	MailStatus_value = map[string]int32{
		"queue":   0,
		"sent":    1,
		"pending": 2,
		"draft":   3,
		"dropped": 4,
	}
)

type MailType int32

const (
	MailType_smtp    MailType = 0
	MailType_inemail MailType = 1
	MailType_chat    MailType = 2
)

// Enum value maps for MailStatus.
var (
	MailType_name = map[int32]string{
		0: "smtp",
		1: "inemail",
		2: "chat",
	}
	MailType_value = map[string]int32{
		"smtp":    0,
		"inemail": 1,
		"chat":    2,
	}
)

type Mail struct {
	Id              string     `gorm:"primaryKey;column:id" json:"id"`
	CommonId        string     `gorm:"column:common_id" json:"common_id"`
	To              string     `gorm:"column:to" json:"to"`     // smtp
	From            string     `gorm:"column:from" json:"from"` // smtp
	OwnerId         int32      `gorm:"column:owner_id" json:"owner_id"`
	UserId          int32      `gorm:"column:user_id" json:"user_id"`          // in-email & chat
	ToUserId        int32      `gorm:"column:to_user_id" json:"to_user_id"`    // in-email & chat
	Subject         string     `gorm:"column:subject" json:"subject"`          // smtp & in-email
	Body            string     `gorm:"column:body" json:"body"`                // smtp & in-email & chat
	ReadAt          *time.Time `gorm:"column:read_at" json:"read_at"`
	Edited          bool       `gorm:"column:edited" json:"edited"`
	Deleted         bool       `gorm:"column:deleted" json:"deleted"`
	Original        bool       `gorm:"column:original" json:"original"` // in the admin panel the broker only need to see one version of the email
	EmailTrackingId string     `gorm:"column:email_tracking_id" json:"email_tracking_id"`
	RepliedToId     string     `gorm:"column:replied_to_id" json:"replied_to_id"`
	ForwardFromId   string     `gorm:"column:forward_from_id" json:"forward_from_id"`
	Replies         []Mail     `gorm:"foreignKey:RepliedToId" json:"replies"`
	Forwards        []Mail     `gorm:"foreignKey:ForwardFromId" json:"forwards"`
	Type            MailType   `gorm:"column:type" json:"type"`
	Status          MailStatus `gorm:"column:status" json:"status"`
	User            User       `gorm:"foreignKey:user_id"`
	CommonModel
}

func (o *Mail) BeforeCreate(tx *gorm.DB) error {
	o.Id = shortuuid.New()
	return nil
}
