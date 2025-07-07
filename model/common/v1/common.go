// Developer: zeelrupapara@gmail.com
// Description: Refined common model types for GreenLync Event-Driven API Gateway
package model

import (
	"time"
)

// ============================================================================
// COMMON STRUCTURES
// ============================================================================

// CommonModel contains common fields for all models
type CommonModel struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	// DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// CORE USER MANAGEMENT
// ============================================================================

// User represents the main user entity for authentication and authorization
type User struct {
	Id           int32  `gorm:"primaryKey;column:id" json:"id"`
	Username     string `gorm:"uniqueIndex;column:username;type:varchar(191)" json:"username"`
	Email        string `gorm:"uniqueIndex;column:email;type:varchar(191)" json:"email"`
	FirstName    string `gorm:"column:first_name;type:varchar(100)" json:"first_name"`
	LastName     string `gorm:"column:last_name;type:varchar(100)" json:"last_name"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	Role         string `gorm:"column:role;type:varchar(50)" json:"role"`
	IsActive     bool   `gorm:"column:is_active;default:true" json:"is_active"`
	LicenseId    string `gorm:"column:license_id;type:varchar(100)" json:"license_id,omitempty"`
	CompanyName  string `gorm:"column:company_name;type:varchar(255)" json:"company_name,omitempty"`
	Phone        string `gorm:"column:phone;type:varchar(20)" json:"phone,omitempty"`
	Address      string `gorm:"column:address;type:text" json:"address,omitempty"`
	CommonModel
}

// UserType represents different types of users
type UserType int32

const (
	UserType_Admin   UserType = 0
	UserType_Manager UserType = 1
	UserType_User    UserType = 2
	UserType_Guest   UserType = 3
	UserType_API     UserType = 4
)

var UserType_name = map[int32]string{
	0: "admin",
	1: "manager",
	2: "user",
	3: "guest",
	4: "api",
}

// ============================================================================
// AUTHENTICATION & AUTHORIZATION
// ============================================================================

// Token represents authentication tokens for JWT-based auth
type Token struct {
	Id           int32     `gorm:"primaryKey;column:id" json:"id"`
	UserId       int32     `gorm:"column:user_id;index" json:"user_id"`
	User         User      `gorm:"foreignKey:UserId" json:"user,omitempty"`
	AccessToken  string    `gorm:"column:access_token;index;type:varchar(255)" json:"access_token"`
	RefreshToken string    `gorm:"column:refresh_token;index;type:varchar(255)" json:"refresh_token"`
	SessionId    string    `gorm:"column:session_id;index;type:varchar(191)" json:"session_id"`
	ExpiresAt    time.Time `gorm:"column:expires_at" json:"expires_at"`
	ExpiresIn    int       `gorm:"column:expires_in" json:"expires_in"`
	TokenType    string    `gorm:"column:token_type;default:Bearer" json:"token_type"`
	Scope        string    `gorm:"column:scope" json:"scope"`
	IpAddress    string    `gorm:"column:ip_address" json:"ip_address"`
	UserAgent    string    `gorm:"column:user_agent" json:"user_agent,omitempty"`
	CommonModel
}

// Permission represents system permissions for RBAC
type Permission struct {
	Id          int32  `gorm:"primaryKey;column:id" json:"id"`
	Name        string `gorm:"uniqueIndex;column:name;type:varchar(191)" json:"name"`
	Description string `gorm:"column:description;type:text" json:"description"`
	Resource    string `gorm:"column:resource;index;type:varchar(191)" json:"resource"`
	Action      string `gorm:"column:action;index;type:varchar(191)" json:"action"`
	IsActive    bool   `gorm:"column:is_active;default:true" json:"is_active"`
	CommonModel
}

// ============================================================================
// EVENT-DRIVEN ARCHITECTURE
// ============================================================================

// Event represents system events for real-time updates and audit trails
type Event struct {
	Id        int32     `gorm:"primaryKey;column:id" json:"id"`
	Type      EventType `gorm:"column:type;index" json:"type"`
	UserId    int32     `gorm:"column:user_id;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserId" json:"user,omitempty"`
	Subject   string    `gorm:"column:subject" json:"subject"`
	Data      string    `gorm:"column:data;type:json" json:"data"`
	Payload   string    `gorm:"column:payload;type:json" json:"payload"`
	Format    string    `gorm:"column:format;default:json" json:"format"`
	SessionId string    `gorm:"column:session_id;index;type:varchar(191)" json:"session_id"`
	IpAddress string    `gorm:"column:ip_address" json:"ip_address,omitempty"`
	Processed bool      `gorm:"column:processed;default:false;index" json:"processed"`
	CommonModel
}

// EventType represents different types of system events
type EventType int32

const (
	// Authentication Events
	EventType_UserLogin       EventType = 0
	EventType_UserLogout      EventType = 1
	EventType_UserRegistered  EventType = 2
	EventType_PasswordChanged EventType = 3
	EventType_SessionExpired  EventType = 4

	// System Events
	EventType_SystemAlert       EventType = 10
	EventType_SystemError       EventType = 11
	EventType_SystemMaintenance EventType = 12
	EventType_ConfigChanged     EventType = 13

	// HTTP Events
	EventType_RequestReceived EventType = 20
	EventType_BadRequest      EventType = 21
	EventType_Unauthorized    EventType = 22
	EventType_Forbidden       EventType = 23
	EventType_NotFound        EventType = 24
	EventType_InternalError   EventType = 25

	// Business Events
	EventType_DataCreated     EventType = 30
	EventType_DataUpdated     EventType = 31
	EventType_DataDeleted     EventType = 32
	EventType_EmailSent       EventType = 33
	EventType_ReportGenerated EventType = 34
	EventType_EmailDraft      EventType = 35
	EventType_EmailOutbox     EventType = 36

	// Compliance Events (for future use)
	EventType_ComplianceCheck  EventType = 40
	EventType_ComplianceUpdate EventType = 41
	EventType_ComplianceAlert  EventType = 42
)

// Event type mappings for serialization
var EventType_name = map[int32]string{
	0:  "user_login",
	1:  "user_logout",
	2:  "user_registered",
	3:  "password_changed",
	4:  "session_expired",
	10: "system_alert",
	11: "system_error",
	12: "system_maintenance",
	13: "config_changed",
	20: "request_received",
	21: "bad_request",
	22: "unauthorized",
	23: "forbidden",
	24: "not_found",
	25: "internal_error",
	30: "data_created",
	31: "data_updated",
	32: "data_deleted",
	33: "email_sent",
	34: "report_generated",
	35: "email_draft",
	36: "email_outbox",
	40: "compliance_check",
	41: "compliance_update",
	42: "compliance_alert",
}

var EventType_value = map[string]int32{
	"user_login":         0,
	"user_logout":        1,
	"user_registered":    2,
	"password_changed":   3,
	"session_expired":    4,
	"system_alert":       10,
	"system_error":       11,
	"system_maintenance": 12,
	"config_changed":     13,
	"request_received":   20,
	"bad_request":        21,
	"unauthorized":       22,
	"forbidden":          23,
	"not_found":          24,
	"internal_error":     25,
	"data_created":       30,
	"data_updated":       31,
	"data_deleted":       32,
	"email_sent":         33,
	"report_generated":   34,
	"email_draft":        35,
	"email_outbox":       36,
	"compliance_check":   40,
	"compliance_update":  41,
	"compliance_alert":   42,
}

// ErrorPayload represents structured error information for events
type ErrorPayload struct {
	Message   string            `json:"message"`
	Code      int               `json:"code"`
	Type      string            `json:"type"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// ============================================================================
// WEBSOCKET & REAL-TIME COMMUNICATION
// ============================================================================

// Message types for WebSocket communication
const (
	TextMessage   = 1
	BinaryMessage = 2
	JsonMessage   = 3
	PingMessage   = "ping"
	PongMessage   = "pong"
)

// ChannelType represents different client connection types
type ChannelType int32

const (
	ChannelType_Web     ChannelType = 0
	ChannelType_Mobile  ChannelType = 1
	ChannelType_Desktop ChannelType = 2
	ChannelType_API     ChannelType = 3
)

var ChannelType_name = map[int32]string{
	0: "web",
	1: "mobile",
	2: "desktop",
	3: "api",
}

// ============================================================================
// CONFIGURATION MANAGEMENT
// ============================================================================

// Config represents system configuration values
type Config struct {
	Id            int32       `gorm:"primaryKey;column:id" json:"id"`
	Key           string      `gorm:"uniqueIndex;column:key;type:varchar(191)" json:"key"`
	Value         string      `gorm:"column:value;type:text" json:"value"`
	ValueType     ValueType   `gorm:"column:value_type;default:0" json:"value_type"`
	Description   string      `gorm:"column:description;type:text" json:"description"`
	ConfigGroupId int32       `gorm:"column:config_group_id;index" json:"config_group_id"`
	ConfigGroup   ConfigGroup `gorm:"foreignKey:ConfigGroupId" json:"config_group,omitempty"`
	IsPublic      bool        `gorm:"column:is_public;default:false" json:"is_public"`
	RecordType    RecordType  `gorm:"column:record_type;default:0;index" json:"record_type"`
	CommonModel
}

// ConfigGroup represents groupings of related configurations
type ConfigGroup struct {
	Id          int32  `gorm:"primaryKey;column:id" json:"id"`
	Name        string `gorm:"uniqueIndex;column:name;type:varchar(191)" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	IsSystem    bool   `gorm:"column:is_system;default:false" json:"is_system"`
	CommonModel
}

// ValueType represents the data type of configuration values
type ValueType int32

const (
	ValueType_String  ValueType = 0
	ValueType_Integer ValueType = 1
	ValueType_Boolean ValueType = 2
	ValueType_Float   ValueType = 3
	ValueType_JSON    ValueType = 4
)

var ValueType_name = map[int32]string{
	0: "string",
	1: "integer",
	2: "boolean",
	3: "float",
	4: "json",
}

// ConvertValueType returns the string representation of a ValueType
func ConvertValueType(vt ValueType) string {
	if name, exists := ValueType_name[int32(vt)]; exists {
		return name
	}
	return "string"
}

// RecordType represents the source/type of a record
type RecordType int32

const (
	RecordType_Seed   RecordType = 0
	RecordType_User   RecordType = 1
	RecordType_System RecordType = 2
	RecordType_Import RecordType = 3
)

// ============================================================================
// OPERATIONS & AUDIT LOGGING
// ============================================================================

// OperationsLog represents audit logs for all system operations
type OperationsLog struct {
	Id          int32  `gorm:"primaryKey;column:id" json:"id"`
	UserId      int32  `gorm:"column:user_id;index" json:"user_id"`
	User        User   `gorm:"foreignKey:UserId" json:"user,omitempty"`
	Action      string `gorm:"column:action;index;type:varchar(191)" json:"action"`
	Resource    string `gorm:"column:resource;index;type:varchar(191)" json:"resource"`
	ResourceId  string `gorm:"column:resource_id;index;type:varchar(191)" json:"resource_id"`
	Method      string `gorm:"column:method" json:"method"`
	URL         string `gorm:"column:url" json:"url"`
	IpAddress   string `gorm:"column:ip_address;index;type:varchar(45)" json:"ip_address"`
	UserAgent   string `gorm:"column:user_agent" json:"user_agent"`
	StatusCode  int    `gorm:"column:status_code;index" json:"status_code"`
	Duration    int64  `gorm:"column:duration" json:"duration"` // milliseconds
	RequestBody string `gorm:"column:request_body;type:text" json:"request_body,omitempty"`
	Response    string `gorm:"column:response;type:text" json:"response,omitempty"`
	SessionId   string `gorm:"column:session_id;index;type:varchar(191)" json:"session_id"`
	CommonModel
}

// ============================================================================
// BUSINESS-SPECIFIC MODELS (EXTENSIBLE)
// ============================================================================

// ComplianceCategory represents compliance categories for business rules
type ComplianceCategory struct {
	Id          int32  `gorm:"primaryKey;column:id" json:"id"`
	Name        string `gorm:"uniqueIndex;column:name;type:varchar(191)" json:"name"`
	Description string `gorm:"column:description;type:text" json:"description"`
	IsMandatory bool   `gorm:"column:is_mandatory;default:false" json:"is_mandatory"`
	IsActive    bool   `gorm:"column:is_active;default:true" json:"is_active"`
	CommonModel
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetEventTypeName returns the string name for an EventType
func GetEventTypeName(eventType EventType) string {
	if name, exists := EventType_name[int32(eventType)]; exists {
		return name
	}
	return "unknown"
}

// GetEventTypeValue returns the EventType for a string name
func GetEventTypeValue(name string) EventType {
	if value, exists := EventType_value[name]; exists {
		return EventType(value)
	}
	return EventType_SystemError
}

// GetChannelTypeName returns the string name for a ChannelType
func GetChannelTypeName(channelType ChannelType) string {
	if name, exists := ChannelType_name[int32(channelType)]; exists {
		return name
	}
	return "unknown"
}

// GetUserTypeName returns the string name for a UserType
func GetUserTypeName(userType UserType) string {
	if name, exists := UserType_name[int32(userType)]; exists {
		return name
	}
	return "unknown"
}
