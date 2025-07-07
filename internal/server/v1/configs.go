// Developer: zeelrupapara@gmail.com
// Description: Simplified configuration management for GreenLync boilerplate
package v1

import (
	"strconv"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ConfigUpdatePayload represents the payload for updating a config
type ConfigUpdatePayload struct {
	Value string `json:"value" validate:"required"`
}

// Simplified UpdateConfig for boilerplate - removes complex trading logic
func (s *HttpServer) UpdateConfig(c *fiber.Ctx) error {
	configId, err := strconv.Atoi(c.Params("config_id"))
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errors.ErrInvalidID)
	}

	payload := &ConfigUpdatePayload{}
	err = c.BodyParser(payload)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	err = s.Validate.Struct(payload)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	config := &model.Config{}
	err = s.DB.First(config, configId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// Simple update - no complex validation for boilerplate
	config.Value = payload.Value

	err = s.DB.Save(config).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// Create a simple config update event for event-driven architecture
	event := &model.Event{
		Type:    model.EventType_SystemAlert, // Using available event type
		UserId:  0, // System event
		Data:    "Config updated",
		Payload: "Config " + c.Params("config_id") + " updated",
		Format:  "json",
	}

	// Save the event to database for audit trail
	s.DB.Create(event)

	// TODO: Publish event to NATS for event-driven processing
	// s.Nats.PublishConfigUpdate(event)

	response := map[string]interface{}{
		"message":   "Configuration updated successfully",
		"config_id": configId,
		"new_value": payload.Value,
	}

	return s.App.HttpResponseOK(c, response)
}

//	@Id				GetAllConfigs
//	@Description	Get All Configs
//	@Tags			Config
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Config
//	@Failure		500	{object}	http.Response
//	@Security		BearerAuth
//	@Router			/api/v1/configs [get]
func (s *HttpServer) GetAllConfigs(c *fiber.Ctx) error {
	configs := &[]*model.Config{}

	err := s.DB.Where("record_type = ?", model.RecordType_Seed).Find(configs).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, configs)
}

//	@Id				GetConfig
//	@Description	Get Config by ID
//	@Tags			Config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.Config
//	@Failure		404	{object}	http.Response
//	@Failure		500	{object}	http.Response
//	@Security		BearerAuth
//	@Param			config_id	path	int	true	"Config ID"
//	@Router			/api/v1/configs/{config_id} [get]
func (s *HttpServer) GetConfig(c *fiber.Ctx) error {
	configId, err := strconv.Atoi(c.Params("config_id"))
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errors.ErrInvalidID)
	}

	config := &model.Config{}

	err = s.DB.Where("record_type = ?", model.RecordType_Seed).First(config, configId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, config)
}

//	@Id				GetConfigsBelongToGroup
//	@Description	Get Configs by Group ID
//	@Tags			Config
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Config
//	@Failure		500	{object}	http.Response
//	@Security		BearerAuth
//	@Param			group_id	path	int	true	"Group ID"
//	@Router			/api/v1/configs/groups/{group_id} [get]
func (s *HttpServer) GetConfigsBelongToGroup(c *fiber.Ctx) error {
	groupId, err := strconv.Atoi(c.Params("group_id"))
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errors.ErrInvalidID)
	}

	configs := &[]*model.Config{}

	err = s.DB.Where("config_group_id = ? AND record_type = ?", groupId, model.RecordType_Seed).Find(configs).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, configs)
}