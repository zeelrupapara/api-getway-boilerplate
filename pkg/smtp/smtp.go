package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"sync"
	"time"

	"greenlync-api-gateway/config"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/db"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/monitor"

	"github.com/go-co-op/gocron"
	"github.com/knadh/smtppool"
)

type SMTP struct {
	Host     string
	Port     int32
	From     string
	Login    string // 💡 New field for SMTP login
	Password string
	DB       *db.MysqlDB
	Log      *logger.Logger
	Corn     *gocron.Scheduler
	Pool     *smtppool.Pool
	sync.RWMutex
}

func NewSmtpClient(cfg *config.Config, log *logger.Logger, db *db.MysqlDB, corn *gocron.Scheduler) (*SMTP, error) {
	smtpClient := &SMTP{
		Host:     cfg.Smtp.SMTP_HOST,
		Port:     cfg.Smtp.SMTP_PORT,
		From:     cfg.Smtp.SMTP_FROM,
		Login:    cfg.Smtp.SMTP_LOGIN, // 💡 Assign login from config
		Password: cfg.Smtp.SMTP_PASSWORD,
		DB:       db,
		Log:      log,
		Corn:     corn,
	}

	var err error
	smtpClient.Pool, err = smtpClient.Connect()
	if err != nil {
		return nil, err
	}

	_, err = corn.Every(1).Second().Do(smtpClient.SendQueuedEmails)
	if err != nil {
		return nil, err
	}

	_, err = corn.Every(10).Second().Do(smtpClient.MonitorSMTPHealth)
	if err != nil {
		return nil, err
	}

	return smtpClient, nil
}

func (s *SMTP) Connect() (*smtppool.Pool, error) {
	var auth smtp.Auth
	// Authentication.
	switch s.Port {
	case 465:
		auth = smtp.PlainAuth("", s.From, s.Password, s.Host)
	case 587:
		auth = smtp.PlainAuth("", s.Login, s.Password, s.Host)
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Host,
	}

	ssl := s.Port == 465

	pool, err := smtppool.New(smtppool.Opt{
		Host:            s.Host,
		Port:            int(s.Port),
		MaxConns:        10,
		Auth:            auth,
		IdleTimeout:     time.Second * 10,
		PoolWaitTimeout: time.Second * 5,
		SSL:             ssl,
		TLSConfig:       tlsconfig,
	})
	return pool, err
}

func (s *SMTP) MonitorSMTPHealth() error {
	var auth smtp.Auth
	switch s.Port {
	case 465:
		auth = smtp.PlainAuth("", s.From, s.Password, s.Host)
	case 587:
		auth = smtp.PlainAuth("", s.Login, s.Password, s.Host)
	}
	smtpAddr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Host,
	}

	var conn net.Conn
	var err error

	if s.Port == 465 {
		conn, err = tls.Dial("tcp", smtpAddr, tlsconfig)
	} else {
		conn, err = net.DialTimeout("tcp", smtpAddr, 5*time.Second)
	}
	if err != nil {
		s.Log.Logger.Errorf("SMTP Health Error [Dial]: %v", err)
		monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
		return nil
	}

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		s.Log.Logger.Errorf("SMTP Health Error [Client]: %v", err)
		monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
		return nil
	}

	if s.Port != 465 {
		if err = client.StartTLS(tlsconfig); err != nil {
			s.Log.Logger.Errorf("SMTP Health Error [STARTTLS]: %v", err)
			monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
			_ = client.Close()
			return nil
		}
	}

	if err = client.Auth(auth); err != nil {
		s.Log.Logger.Errorf("SMTP Health Error [AUTH]: %v", err)
		monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
		_ = client.Close()
		return nil
	}

	if err = client.Noop(); err != nil {
		s.Log.Logger.Errorf("SMTP Health Error [NOOP]: %v", err)
		monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
	}

	_ = client.Quit()
	monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_running)
	return nil
}

func (s *SMTP) QueueEmail(to, subject, body string) error {
	mail := &model.Mail{
		Subject: subject,
		Body:    body,
		From:    s.From,
		To:      to,
		Status:  model.MailStatus_queue,
		Type:    model.MailType_smtp,
	}

	return s.DB.DB.Create(mail).Error
}

func (s *SMTP) SendEmail(email *model.Mail) error {
	e := smtppool.Email{
		From:    email.From,
		To:      []string{email.To},
		Subject: email.Subject,
		HTML:    []byte(email.Body),
	}

	err := s.Pool.Send(e)
	if err != nil {
		s.Log.Logger.Errorf("SMTP Send Failed: %v", err)
		_ = s.DB.DB.Model(email).UpdateColumn("status", model.MailStatus_dropped).Error
		return err
	}

	s.Log.Logger.Infof("Email sent to: %s", email.To)
	_ = s.DB.DB.Model(email).UpdateColumn("status", model.MailStatus_sent).Error
	return nil
}

func (s *SMTP) SendQueuedEmails() {
	s.Lock()
	defer s.Unlock()

	var mail []*model.Mail
	err := s.DB.DB.Where("status = ? AND type = ?", model.MailStatus_queue, model.MailType_smtp).Find(&mail).Error
	if err != nil {
		s.Log.Logger.Errorf("SMTP Queue Fetch Error: %v", err)
	}

	for _, m := range mail {
		err = s.DB.DB.Model(m).UpdateColumn("status", model.MailStatus_sent).Error
		if err != nil {
			s.Log.Logger.Error("Error updating mail status to sent")
		}
		go s.SendEmail(m)
	}
}
