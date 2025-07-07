package v1

import (
	"fmt"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

//	@Id				GetAllOperations
//	@Description	Get All Operations
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Param			account_id	query		int		false	"search by acccount_id"
//	@Param			page		query		int		false	"page number"
//	@Param			limit		query		int		false	"limit number"
//	@Param			from		query		string	false	"from date"
//	@Param			to			query		string	false	"to date"
//	@Param			sort_by		query		string	false	"sort by"
//	@Success		200			{array}		model.OperationsLog
//	@Failure		500			{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/system/operations [get]
func (s *HttpServer) GetAllOperations(c *fiber.Ctx) error {
	// request filters (date range, page number and page limit) if exists
	query, err := utils.QueryFilter(c)
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// accountId
	accountId := c.QueryInt("account_id", 0)
	if accountId != 0 {
		if query.IsEmpty {
			query.QueryString += fmt.Sprintf("account_id = %d", accountId)
		} else {
			query.QueryString += fmt.Sprintf(" AND account_id = %d", accountId)
		}
	}

	operations := []model.OperationsLog{}
	err = s.DB.
		Where(query.QueryString).
		Offset(query.Page * query.Limit).
		Limit(query.Limit).
		Order(query.SortBy).
		Find(&operations).
		Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, operations)
}

//	@Id				GetOperation
//	@Description	Get Operation
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.OperationsLog
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			operation_id	path	int	true	"Operation ID"
//	@Router			/api/v1/system/operations/{operation_id} [get]
func (s *HttpServer) GetOperation(c *fiber.Ctx) error {
	operationId, err := c.ParamsInt("operation_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	operation := &model.OperationsLog{}
	err = s.DB.First(operation, operationId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, operation)
}

//	@Id				DeleteOperation
//	@Description	Delete Operation
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Param			operation_id	path	int	true	"Operation ID"
//	@Router			/api/v1/system/operations/{operation_id} [DELETE]
func (s *HttpServer) DeleteOperation(c *fiber.Ctx) error {
	operationId, err := c.ParamsInt("operation_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	operation := &model.OperationsLog{}
	err = s.DB.First(operation, operationId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	err = s.DB.Delete(operation).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// TODO: Implement operation deletion events
	// operationPayload, _ := json.Marshal(operation)
	// event := &model.Event{
	// 	Type:    model.EventType_DataDeleted,
	// 	Payload: string(operationPayload),
	// 	Format:  "json",
	// }
	// TODO: Implement WebSocket publishing
	// err = s.PublishWS(subject, event)

	// TODO: Implement journal event handling for boilerplate
	// Simple operation deletion completed

	return s.App.HttpResponseNoContent(c)
}

//	@Id				DeleteAllOperations
//	@Description	Delete All Operations
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/api/v1/system/operations [DELETE]
func (s *HttpServer) DeleteAllOperations(c *fiber.Ctx) error {
	err := s.DB.Delete(&model.OperationsLog{}).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// TODO: Implement bulk delete events
	// event := &model.Event{
	// 	Type:    model.EventType_DataDeleted,
	// 	Payload: "all operations have been deleted",
	// 	Format:  "json",
	// }
	// TODO: Implement WebSocket publishing
	// err = s.PublishWS(subject, event)

	// TODO
	// broadcast to all clients that the journals of theirs have been deleted

	return s.App.HttpResponseNoContent(c)
}

func (s *HttpServer) queueSystemOperationLog(operation *model.OperationsLog) {
	s.operationCh <- operation
}

func (s *HttpServer) writeSystemOperationsLogs() {
	var operation *model.OperationsLog
	for {
		// unmarshel data
		operation = <-s.operationCh

		// create record
		err := s.DB.Create(operation).Error
		if err != nil {
			s.Log.Logger.Errorf("error creating operation log for userId %d action %s",
				operation.UserId, operation.Action)
		}

		// TODO: Implement operation creation events
		// operationPayload, _ := json.Marshal(operation)
		// event := &model.Event{
		// 	Type:    model.EventType_DataCreated,
		// 	Payload: string(operationPayload),
		// 	Format:  "json",
		// }
		// TODO: Implement WebSocket publishing
		// err = s.PublishWS(subject, event)

		// TODO: Implement journal events for boilerplate
		// Simple operation logging completed
	}
}
