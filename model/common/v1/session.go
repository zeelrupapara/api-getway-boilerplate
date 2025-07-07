package model

import (
	"time"

	"github.com/lithammer/shortuuid/v4"
	"gorm.io/gorm"
)

type Session struct {
	Id         string     `gorm:"primaryKey;column:id" json:"id"`
	UserId     int32      `gorm:"column:user_id" json:"user_id"`
	UserAgent  string     `gorm:"column:user_agent" json:"user_agent"`
	User       User       `gorm:"foreignKey:UserId" json:"user"`
	IpAddress  string     `gorm:"column:ip_address" json:"ip_address"`
	SessionId  string     `gorm:"index;column:session_id" json:"session_id"`
	StartedAt  time.Time  `gorm:"column:started_at" json:"started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finished_at"`
	Scope      string     `gorm:"column:scope" json:"scope"`
	CommonModel
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	s.Id = shortuuid.New()
	s.StartedAt = time.Now()
	return nil
}
