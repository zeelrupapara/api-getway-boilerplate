// Developer: zeelrupapara@gmail.com
// Description: Email management for GreenLync boilerplate
package v1

import (
	"fmt"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/pkg/shortuuid"
	"greenlync-api-gateway/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

//	@Id				GetAllInEmails
//	@Description	Get All InEmails
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails [get]
func (s *HttpServer) GetAllInEmails(c *fiber.Ctx) error {
	mail := []*model.Mail{}

	err := s.DB.Where("type = ?", model.MailType_notification).Find(&mail).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, mail)
}

//	@Id				GetAllAccountInEmails
//	@Description	Get Account InEmails Inbox
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails/me/inbox [get]
func (s *HttpServer) GetAllAccountInEmails(c *fiber.Ctx) error {
	mail := []*model.Mail{}

	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	// email_tracking_id
	err := s.DB.
		Where("type = ? AND status = ? AND owner_id = ? AND to_account_id = ? AND deleted = false",
			model.MailType_notification, model.MailStatus_sent, client.ClientId, client.ClientId).
		Order("created_at DESC").Find(&mail).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// temp
	m := make(map[string]struct{})
	filterdInbox := []*model.Mail{}
	for i := range mail {
		if _, ok := m[mail[i].EmailTrackingId]; !ok {
			filterdInbox = append(filterdInbox, mail[i])
			if mail[i].EmailTrackingId != "" {
				m[mail[i].EmailTrackingId] = struct{}{}
			}
		}
	}

	return s.App.HttpResponseOK(c, filterdInbox)
}

//	@Id				GetAllAccountOutEmails
//	@Description	Get Account InEmails Mail
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails/me/outbox [get]
func (s *HttpServer) GetAllAccountOutEmails(c *fiber.Ctx) error {
	mail := []*model.Mail{}

	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	err := s.DB.
		Where("type = ? AND status = ? AND owner_id = ? AND account_id = ?  AND deleted = false",
			model.MailType_notification, model.MailStatus_sent, client.ClientId, client.ClientId).
		Order("created_at DESC").Find(&mail).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// temp
	m := make(map[string]struct{})
	filterdInbox := []*model.Mail{}
	for i := range mail {
		if _, ok := m[mail[i].EmailTrackingId]; !ok {
			filterdInbox = append(filterdInbox, mail[i])
			if mail[i].EmailTrackingId != "" {
				m[mail[i].EmailTrackingId] = struct{}{}
			}
		}
	}

	return s.App.HttpResponseOK(c, filterdInbox)
}

//	@Id				GetAccountDraftEmails
//	@Description	Get Account InEmails Draft
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails/me/draft [get]
func (s *HttpServer) GetAccountDraftEmails(c *fiber.Ctx) error {
	mail := []*model.Mail{}

	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	err := s.DB.
		Where("type = ? AND status = ? AND owner_id = ? AND account_id = ? AND deleted = false",
			model.MailType_notification, model.MailStatus_draft, client.ClientId, client.ClientId).
		Order("created_at DESC").Find(&mail).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// temp
	m := make(map[string]struct{})
	filterdInbox := []*model.Mail{}
	for i := range mail {
		if _, ok := m[mail[i].EmailTrackingId]; !ok {
			filterdInbox = append(filterdInbox, mail[i])
			if mail[i].EmailTrackingId != "" {
				m[mail[i].EmailTrackingId] = struct{}{}
			}
		}
	}

	return s.App.HttpResponseOK(c, filterdInbox)
}

//	@Id				GetAccountBinEmails
//	@Description	Get Account InEmails Bin
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails/me/bin [get]
func (s *HttpServer) GetAccountBinEmails(c *fiber.Ctx) error {
	mail := []*model.Mail{}

	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	err := s.DB.
		Where("type = ? AND owner_id = ? AND deleted = true",
			model.MailType_notification, client.ClientId).
		// Where("common_id IS NULL OR common_id NOT IN (SELECT DISTINCT replied_to_id FROM vfxMail WHERE replied_to_id IS NOT NULL)").
		Order("created_at DESC").Find(&mail).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// temp
	m := make(map[string]struct{})
	filterdInbox := []*model.Mail{}
	for i := range mail {
		if _, ok := m[mail[i].EmailTrackingId]; !ok {
			filterdInbox = append(filterdInbox, mail[i])
			if mail[i].EmailTrackingId != "" {
				m[mail[i].EmailTrackingId] = struct{}{}
			}
		}
	}

	return s.App.HttpResponseOK(c, filterdInbox)
}

//	@Id				GetInEmail
//	@Description	Get InEmail
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID"
//	@Router			/api/v1/emails/{id} [get]
func (s *HttpServer) GetInEmail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("id %s", errors.ErrRequiredParams))
	}

	mail := &model.Mail{}
	err := s.DB.Where(&model.Mail{Id: id, Type: model.MailType_notification}).Preload("Replies").First(mail).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, mail)
}

//	@Id				GetAccountInEmail
//	@Description	Get Account InEmail Details
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Mail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			tracking_id	path	string	true	"Tracking ID"
//	@Router			/api/v1/emails/me/{tracking_id} [get]
func (s *HttpServer) GetAccountInEmail(c *fiber.Ctx) error {
	trackingId := c.Params("tracking_id")
	if trackingId == "" {
		return s.App.HttpResponseInternalServerErrorRequest(c, fmt.Errorf("tracking_id %s", errors.ErrRequiredParams))
	}

	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	mail := &[]*model.Mail{}
	err := s.DB.Where("type = ? AND status = ? AND email_tracking_id = ? AND owner_id = ?  AND deleted = false",
		model.MailType_notification, model.MailStatus_sent, trackingId, client.ClientId).Order("created_at DESC").Find(mail).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if len(*mail) == 0 {
		return s.App.HttpResponseNotFound(c, gorm.ErrRecordNotFound)
	}

	return s.App.HttpResponseOK(c, mail)
}

type InEmail struct {
	ToUserId []int32          `json:"to" validate:"required"`
	ReplyTo  string           `json:"reply_to"`
	Subject  string           `json:"subject" validate:"required"`
	Body     string           `json:"body" validate:"required"`
	Status   model.MailStatus `json:"status" validate:"required"`
}

//	@Id				CreateInEmail
//	@Description	Create InEmail
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	v1.InEmail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			body	body	v1.InEmail	true	"InEmail Request Body"
//	@Router			/api/v1/emails [post]
func (s *HttpServer) CreateInEmail(c *fiber.Ctx) error {
	data := &InEmail{}
	err := c.BodyParser(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	err = s.Validate.Struct(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	cfg, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	tx := s.DB.Begin()

	trackingId := ""
	if data.ReplyTo != "" {
		reply := &model.Mail{}
		err = tx.Where("common_id = ?", data.ReplyTo).First(reply).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		if reply.EmailTrackingId == "" {
			trackingId = shortuuid.New()
			err = tx.Model(&model.Mail{}).Where("common_id = ?", data.ReplyTo).Update("email_tracking_id", trackingId).Error
			if err != nil {
				tx.Rollback()
				if err == gorm.ErrRecordNotFound {
					return s.App.HttpResponseBadRequest(c, err)
				}
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
		} else {
			trackingId = reply.EmailTrackingId
		}
	}

	emails := []*model.Mail{}
	sendingEmails := []*model.Mail{}
	commonId := shortuuid.New()
	for i := range data.ToUserId {
		originalEmail := &model.Mail{
			CommonId:        commonId,
			UserId:          cfg.ClientId,
			ToUserId:        data.ToUserId[i],
			Subject:         data.Subject,
			Body:            data.Body,
			Status:          data.Status,
			OwnerId:         cfg.ClientId,
			Original:        true,
			Type:            model.MailType_notification,
			EmailTrackingId: trackingId,
		}
		toEmail := *originalEmail
		toEmail.OwnerId = toEmail.ToUserId
		toEmail.Original = false
		emails = append(emails, originalEmail, &toEmail)
		sendingEmails = append(sendingEmails, &toEmail)

		err = tx.Save(&emails).Error
		if err != nil {
			tx.Rollback()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	// Convert email to JSON payload
	emailPayload, _ := json.Marshal(emails[0])
	event := &model.Event{
		Type:    model.EventType_EmailDraft,
		Payload: string(emailPayload),
		Format:  "json",
	}
	if data.Status == model.MailStatus_sent {
		event.Type = model.EventType_EmailOutbox
	}
	// TODO: Implement WebSocket publishing for event-driven architecture
	// err = s.PublishWS(model.SubjectAccountEmails(cfg.ClientId), event)
	// if err != nil {
	// 	tx.Rollback()
	// 	return s.App.HttpResponseInternalServerErrorRequest(c, err)
	// }

	if data.Status == model.MailStatus_sent {
		event.Type = model.EventType_EmailSent
		for i := range sendingEmails {
			emailPayload, _ := json.Marshal(sendingEmails[i])
			event.Payload = string(emailPayload)
			// TODO: Implement WebSocket publishing
			// err = s.PublishWS(model.SubjectAccountEmails(sendingEmails[i].OwnerId), event)
			// if err != nil {
			// 	s.Log.Logger.Error(err)
			// } // we can't really undo the already sent messages
		}
	}

	tx.Commit()

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "create_email",
		Resource:  "email",
		UserId:    cfg.ClientId,
		Method:    "POST",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseOK(c, emails[0])
}

//	@Id				UpdateInEmail
//	@Description	Update Account InEmail if the status is draft
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.InEmail
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			id		path	string		true	"Email ID"
//	@Param			body	body	v1.InEmail	true	"InEmail Request Body"
//	@Router			/api/v1/emails/{id} [put]
func (s *HttpServer) UpdateInEmail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("id %s", errors.ErrRequiredParams))
	}

	cfg, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	data := &InEmail{}
	err := c.BodyParser(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	err = s.Validate.Struct(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	draftEmail := &model.Mail{}
	err = s.DB.Where("type = ? AND id = ? AND account_id = ?",
		model.MailType_notification, id, cfg.ClientId).First(draftEmail).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if draftEmail.Status != model.MailStatus_draft {
		return s.App.HttpResponseBadRequest(c, errors.ErrEmailHasAlreadyBeenSent)
	}

	trackingId := draftEmail.EmailTrackingId
	if data.ReplyTo != "" {
		reply := &model.Mail{}
		err = s.DB.Where("common_id = ?", data.ReplyTo).First(reply).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if reply.EmailTrackingId == "" {
			trackingId = shortuuid.New()
		} else {
			trackingId = reply.EmailTrackingId
		}
	}

	tx := s.DB.Begin()
	err = tx.Where("type = ? AND common_id = ? AND account_id = ?",
		model.MailType_notification, draftEmail.CommonId, cfg.ClientId).Delete(&model.Mail{}).Error
	if err != nil {
		tx.Rollback()
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	emails := []*model.Mail{}
	sendingEmails := []*model.Mail{}
	for i := range data.ToUserId {
		originalEmail := &model.Mail{
			CommonId:        draftEmail.CommonId,
			UserId:          cfg.ClientId,
			ToUserId:        data.ToUserId[i],
			Subject:         data.Subject,
			Body:            data.Body,
			Status:          data.Status,
			OwnerId:         cfg.ClientId,
			Original:        true,
			Type:            model.MailType_notification,
			EmailTrackingId: trackingId,
		}
		toEmail := *originalEmail
		toEmail.OwnerId = toEmail.ToUserId
		toEmail.Original = false
		emails = append(emails, originalEmail, &toEmail)
		sendingEmails = append(sendingEmails, &toEmail)
	}

	err = tx.Save(&emails).Error
	if err != nil {
		tx.Rollback()
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// Convert email to JSON payload
	emailPayload, _ := json.Marshal(emails[0])
	event := &model.Event{
		Type:    model.EventType_EmailDraft,
		Payload: string(emailPayload),
		Format:  "json",
	}
	if data.Status == model.MailStatus_sent {
		event.Type = model.EventType_EmailOutbox
	}
	// TODO: Implement WebSocket publishing for event-driven architecture
	// err = s.PublishWS(model.SubjectAccountEmails(cfg.ClientId), event)
	// if err != nil {
	// 	tx.Rollback()
	// 	return s.App.HttpResponseInternalServerErrorRequest(c, err)
	// }

	if data.Status == model.MailStatus_sent {
		event.Type = model.EventType_EmailSent
		for i := range sendingEmails {
			emailPayload, _ := json.Marshal(sendingEmails[i])
			event.Payload = string(emailPayload)
			// TODO: Implement WebSocket publishing
			// err = s.PublishWS(model.SubjectAccountEmails(sendingEmails[i].OwnerId), event)
			// if err != nil {
			// 	s.Log.Logger.Error(err)
			// } // we can't really undo the already sent messages
		}
	}

	tx.Commit()

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "update_email",
		Resource:  "email",
		UserId:    cfg.ClientId,
		Method:    "PUT",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseOK(c, emails[0])
}

//	@Id				DeleteInEmail
//	@Description	move Account InEmail to bin then delete it if it's already in the bin
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID"
//	@Router			/api/v1/emails/{id} [DELETE]
func (s *HttpServer) DeleteInEmail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("id %s", errors.ErrRequiredParams))
	}

	cfg, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	mail := &model.Mail{}
	err := s.DB.Where("type = ? AND owner_id = ? AND id = ?", model.MailType_notification, cfg.ClientId, id).First(mail).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// publish to account
	mailPayload, _ := json.Marshal(mail)
	event := &model.Event{
		Type:    model.EventType_DataDeleted,
		Payload: string(mailPayload),
		Format:  "json",
	}

	tx := s.DB.Begin()

	if mail.Status == model.MailStatus_draft {
		err = tx.Where("type = ? AND common_id = ? AND user_id = ?",
			model.MailType_notification, mail.CommonId, cfg.ClientId).Delete(&model.Mail{}).Error
		if err != nil {
			tx.Rollback()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	} else {
		if mail.Deleted {
			err = tx.Delete(mail).Error
			if err != nil {
				tx.Rollback()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
		} else {
			err = tx.Model(mail).UpdateColumn("deleted", true).Error
			if err != nil {
				tx.Rollback()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
			event.Type = model.EventType_DataUpdated
		}
	}

	// TODO: Implement WebSocket publishing for event-driven architecture
	// err = s.PublishWS(model.SubjectAccountEmails(cfg.ClientId), event)
	// if err != nil {
	// 	tx.Rollback()
	// 	return s.App.HttpResponseInternalServerErrorRequest(c, err)
	// }

	tx.Commit()

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "delete_email",
		Resource:  "email",
		UserId:    cfg.ClientId,
		Method:    "DELETE",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseNoContent(c)
}

//	@Id				TestSMTPEmail
//	@Description	Test SMTP
//	@Tags			Emails
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/emails/test [post]
func (s *HttpServer) TestSMTPEmail(c *fiber.Ctx) error {
	email := ""
	err := c.BodyParser(&email)
	if err != nil {
		s.App.HttpResponseBadRequest(c, err)
	}

	cfg, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	s.Smtp.QueueEmail(email, "Dear All", "<html><body><h1>Dear Customers I would like to tell you that vertex12 is almost live</h1></body></html>")

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "send_test_email",
		Resource:  "email",
		UserId:    cfg.ClientId,
		Method:    "POST",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseNoContent(c)
}
