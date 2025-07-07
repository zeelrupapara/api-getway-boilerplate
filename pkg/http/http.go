// Developer: Saif Hamdan
package http

import (
	"fmt"
	"time"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/logger"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

// Locals constants
const (
	LocalsAllowed = "allowed"
	LocalsClient  = "client"
	LocalsToken   = "token"
	LocalsDevice  = "device"
	LocalsOs      = "os"
	LocalsChannel = "channel"
)

const (
	StatusBadRequest          = fiber.StatusBadRequest
	StatusUnauthorized        = fiber.StatusUnauthorized
	StatusForbidden           = fiber.StatusForbidden
	StatusNotFound            = fiber.StatusNotFound
	StatusInternalServerError = fiber.StatusInternalServerError
	StatusOK                  = fiber.StatusOK
	StatusCreated             = fiber.StatusCreated
	StatusNoContent           = fiber.StatusNoContent
)

const (
	ErrBadRequest          = "Bad request"
	ErrInternalServerError = "Internal server error"
	ErrAlreadyExists       = "Already exists"
	ErrNotFound            = "Not Found"
	ErrUnauthorized        = "Unauthorized"
	ErrForbidden           = "Forbidden"
	ErrBadQueryParams      = "Invalid query params"
	ErrRequestTimeout      = "Request Timeout"
	ErrEndpointNotFound    = "The endpoint you requested doesn't exist on server"
)

type App struct {
	// fiber app instence
	*fiber.App
	// logger
	Log *logger.Logger
}

func NewApp(log *logger.Logger) *App {
	newapp := fiber.New(fiber.Config{
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		EnableTrustedProxyCheck: true,
	})

	return &App{
		App: newapp,
		Log: log,
	}
}

type HttpResponse struct {
	// Response flag indicates whether the HTTP request was successful or not
	Success bool `json:"success"`
	// Http status Code
	Code int `json:"code"`
	// if the request were successful the data will be saved here
	Data interface{} `json:"data"`
	// Generic General Error Message defined in the system
	Error string `json:"error"`
	// More detailed error message indicates why the request was unsuccessful
	Message string `json:"message"`
}

type WSResponse struct {
	// Event Reason
	EventReason string
	// Response flag indicates whether the HTTP request was successful or not
	Success bool `json:"success"`
	// Http status Code
	Code int `json:"code"`
	// if the request were successful the data will be saved here
	Data interface{} `json:"data"`
	// Generic General Error Message defined in the system
	Error string `json:"error"`
	// More detailed error message indicates why the request was unsuccessful
	Message string `json:"message"`
}

// http 200 ok http response
func (a *App) HttpResponseOK(c *fiber.Ctx, data interface{}) error {
	return c.Status(StatusOK).JSON(
		&HttpResponse{
			Success: true,
			Code:    StatusOK,
			Data:    data,
			Error:   "",
			Message: "",
		})
}

// http 201 created http response
func (a *App) HttpResponseCreated(c *fiber.Ctx, data interface{}) error {
	return c.Status(StatusCreated).JSON(
		&HttpResponse{
			Success: true,
			Code:    StatusCreated,
			Data:    data,
			Error:   "",
			Message: "",
		})
}

// http 204 no content http response
func (a *App) HttpResponseNoContent(c *fiber.Ctx) error {
	return c.Status(StatusNoContent).JSON(
		&HttpResponse{
			Success: true,
			Code:    StatusNoContent,
			Data:    nil,
			Error:   "",
			Message: "",
		})
}

// http 400 bad request http response
func (a *App) HttpResponseBadRequest(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusBadRequest).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusBadRequest,
			Data:    nil,
			Error:   ErrBadRequest,
			Message: message.Error(),
		})
}

// http 400 bad query params http response
func (a *App) HttpResponseBadQueryParams(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusBadRequest).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusBadRequest,
			Data:    nil,
			Error:   ErrBadQueryParams,
			Message: message.Error(),
		})
}

// http 404 not found http response
func (a *App) HttpResponseNotFound(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusNotFound).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusNotFound,
			Data:    nil,
			Error:   ErrNotFound,
			Message: message.Error(),
		})
}

// http 500 internal server error response
func (a *App) HttpResponseInternalServerErrorRequest(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusInternalServerError).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusInternalServerError,
			Data:    nil,
			Error:   ErrInternalServerError,
			Message: message.Error(),
		})
}

// http 403 The client does not have access rights to the content;
// that is, it is unauthorized, so the server is refusing to give the requested resource
func (a *App) HttpResponseForbidden(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusForbidden).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusForbidden,
			Data:    nil,
			Error:   ErrForbidden,
			Message: message.Error(),
		})
}

// http 401 the client must authenticate itself to get the requested response
func (a *App) HttpResponseUnauthorized(c *fiber.Ctx, message error) error {
	a.Log.Logger.Error(message.Error())
	return c.Status(StatusUnauthorized).JSON(
		&HttpResponse{
			Success: false,
			Code:    StatusUnauthorized,
			Data:    nil,
			Error:   ErrUnauthorized,
			Message: message.Error(),
		})
}

// http 200 retrieve File response
func (a *App) HttpResponseFile(c *fiber.Ctx, file []byte) error {
	return c.Status(fiber.StatusOK).Send(file)
}

// WS 200 ok http response
func (a *App) WSResponseOK(event model.EventType, data interface{}) *model.Event {
	dataStr := ""
	if data != nil {
		if str, ok := data.(string); ok {
			dataStr = str
		} else {
			dataStr = fmt.Sprintf("%v", data)
		}
	}
	return &model.Event{
		Type:    event,
		Payload: dataStr,
	}
}

// WS 400 bad request http response
func (a *App) WSResponseBadRequest(event model.EventType, message error) *model.Event {
	a.Log.Logger.Error(message.Error())
	errorPayload := model.ErrorPayload{
		Message:   message.Error(),
		Code:      400,
		Type:      "bad_request",
		Timestamp: time.Now(),
	}
	payloadStr, _ := json.Marshal(errorPayload)
	return &model.Event{
		Type:    model.EventType_BadRequest,
		Payload: string(payloadStr),
	}
}

// WS 500 internal server error response
func (a *App) WSResponseInternalServerErrorRequest(event model.EventType, message error) *model.Event {
	a.Log.Logger.Error(message.Error())
	errorPayload := model.ErrorPayload{
		Message:   message.Error(),
		Code:      500,
		Type:      "internal_error",
		Timestamp: time.Now(),
	}
	payloadStr, _ := json.Marshal(errorPayload)
	return &model.Event{
		Type:    model.EventType_InternalError,
		Payload: string(payloadStr),
	}
}

// WS 404 not found http response
func (a *App) WSResponseNotFound(event model.EventType, message error) *model.Event {
	a.Log.Logger.Error(message.Error())
	errorPayload := model.ErrorPayload{
		Message:   message.Error(),
		Code:      404,
		Type:      "not_found",
		Timestamp: time.Now(),
	}
	payloadStr, _ := json.Marshal(errorPayload)
	return &model.Event{
		Type:    model.EventType_NotFound,
		Payload: string(payloadStr),
	}
}

// WS 403 The client does not have access rights to the content;
// that is, it is unauthorized, so the server is refusing to give the requested resource
func (a *App) WSResponseForbidden(event model.EventType, message error) *model.Event {
	a.Log.Logger.Error(message.Error())
	errorPayload := model.ErrorPayload{
		Message:   message.Error(),
		Code:      403,
		Type:      "forbidden",
		Timestamp: time.Now(),
	}
	payloadStr, _ := json.Marshal(errorPayload)
	return &model.Event{
		Type:    model.EventType_Forbidden,
		Payload: string(payloadStr),
	}
}

// ws 401 the client must authenticate itself to get the requested response
func (a *App) WSResponseUnauthorized(event model.EventType, message error) *model.Event {
	a.Log.Logger.Error(message.Error())
	errorPayload := model.ErrorPayload{
		Message:   message.Error(),
		Code:      401,
		Type:      "unauthorized",
		Timestamp: time.Now(),
	}
	payloadStr, _ := json.Marshal(errorPayload)
	return &model.Event{
		Type:    model.EventType_Unauthorized,
		Payload: string(payloadStr),
	}
}
