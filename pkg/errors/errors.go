// By Emran A. Hamdan, Lead Architect & Saif Hamdan, Team Lead
package errors

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

const (
	BadRequest                      = "bad request"
	InternalServerError             = "internal server error"
	AlreadyExists                   = "already exists"
	NotFound                        = "not Found"
	Unauthorized                    = "unauthorized"
	Unauthenticated                 = "unauthenticated"
	Forbidden                       = "forbidden"
	BadQueryParams                  = "invalid query params"
	TooManyRequests                 = "too many request"
	InvalidField                    = "invalid field"
	RequiredField                   = "required field"
	RequiredParams                  = "param is required"
	RequestTimeout                  = "Request Timeout"
	MissingAuthoirzationHeader      = "missing Authorization header"
	BasicAuth                       = "authentication error: please provide valid basic authentication credentials in the 'Authorization' header"
	BearerToken                     = "authentication error: please provide a valid bearer token in the 'Authorization' header"
	InvalidToken                    = "expired or invalid token"
	InvalidSession                  = "expired or invalid session"
	SessionUsed                     = "session is already used"
	UnauthorizedToAccessResource    = "unauthorized to access this resource"
	EndpointNotFound                = "the endpoint you requested doesn't exist on server"
	CouldNotParseClientCfg          = "couldn't parse client stored config"
	DuplicateField                  = "duplicate entry for key"
	IncorrectAccountType            = "incorrect account type"
	IncorrectTransactionType        = "incorrect transaction type"
	AmountMoreThanZero              = "amount should be more than 0"
	AmountMoreOrLessThanZero        = "amount should be more or less than 0"
	AmountMoreOrLessThanZeroCredit  = "amount should be more than 0 and less than or equal to credit"
	RecordCouldntBeUpdatedOrDeleted = "this record couldn't be deleted or updated"
	EmailHasAlreadyBeenSent         = "the email has been already sent"
	DeleteWhileNotEmpty             = "you can't delete a record while it's not empty"
	MarketAlreadyStarted            = "market feed has been already started"
	MarketAlreadyStoped             = "market feed has been already stopped"
	InvalidID                       = "invalid ID parameter"
)

var (
	ErrBadRequest                      = errors.New(BadRequest)
	ErrUnauthorized                    = errors.New(Unauthorized)
	ErrUnauthenticated                 = errors.New(Unauthenticated)
	ErrEndpointNotFound                = errors.New(EndpointNotFound)
	ErrInternalServerError             = errors.New(InternalServerError)
	ErrTooManyRequests                 = errors.New(TooManyRequests)
	ErrRequiredParams                  = errors.New(RequiredParams)
	ErrInvalidField                    = errors.New(InvalidField)
	ErrInvalidToken                    = errors.New(InvalidToken)
	ErrInvalidSession                  = errors.New(InvalidSession)
	ErrSessionUsed                     = errors.New(SessionUsed)
	ErrInvalidBasicAuth                = errors.New(BasicAuth)
	ErrInvalidBearerToken              = errors.New(BearerToken)
	ErrCouldNotParseClientCfg          = errors.New(CouldNotParseClientCfg)
	ErrIncorrectAccountType            = errors.New(IncorrectAccountType)
	ErrIncorrectTransactionType        = errors.New(IncorrectTransactionType)
	ErrAmountMoreThanZero              = errors.New(AmountMoreThanZero)
	ErrAmountMoreOrLessThanZero        = errors.New(AmountMoreOrLessThanZero)
	ErrAmountMoreOrLessThanZeroCredit  = errors.New(AmountMoreOrLessThanZeroCredit)
	ErrRecordCouldntBeUpdatedOrDeleted = errors.New(RecordCouldntBeUpdatedOrDeleted)
	ErrEmailHasAlreadyBeenSent         = errors.New(EmailHasAlreadyBeenSent)
	ErrDeleteWhileNotEmpty             = errors.New(DeleteWhileNotEmpty)
	ErrMarketAlreadyStarted            = errors.New(MarketAlreadyStarted)
	ErrMarketAlreadyStoped             = errors.New(MarketAlreadyStoped)
	ErrInvalidID                       = errors.New(InvalidID)
	ErrUnauthorizedToAccessResource    = errors.New(UnauthorizedToAccessResource)
	ErrMissingAuthoirzationHeader      = errors.New(MissingAuthoirzationHeader)
)

type HttpErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

func BadQueryParamsHttpResponse(message string) *HttpErrorResponse {
	return &HttpErrorResponse{
		StatusCode: fiber.StatusBadRequest,
		Error:      BadQueryParams,
		Message:    message,
	}
}

func NotFoundHttpResponse(message string) *HttpErrorResponse {
	return &HttpErrorResponse{
		StatusCode: fiber.StatusNotFound,
		Error:      NotFound,
		Message:    message,
	}
}

func BadRequestHttpResponse(message string) *HttpErrorResponse {
	return &HttpErrorResponse{
		StatusCode: fiber.StatusBadRequest,
		Error:      BadRequest,
		Message:    message,
	}
}

func InternalServerErrorRequestHttpResponse(message string) *HttpErrorResponse {
	return &HttpErrorResponse{
		StatusCode: fiber.StatusInternalServerError,
		Error:      InternalServerError,
		Message:    message,
	}
}

func UnauthorizedHttpResponse(message string) *HttpErrorResponse {
	return &HttpErrorResponse{
		StatusCode: fiber.StatusUnauthorized,
		Error:      Unauthorized,
		Message:    message,
	}
}

func NewDoesntExistError(subject string) error {
	return errors.New(subject + " Does not exist")
}
