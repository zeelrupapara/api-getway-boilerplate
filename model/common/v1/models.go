package v1

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Cannabis User Roles
type UserRole string

const (
	RoleCustomer          UserRole = "customer"
	RoleBudtender         UserRole = "budtender"
	RoleDispensaryManager UserRole = "dispensary_manager"
	RoleBrandPartner      UserRole = "brand_partner"
	RoleSystemAdmin       UserRole = "system_admin"
)

// Cannabis Compliance Status
type ComplianceStatus string

const (
	ComplianceStatusPending   ComplianceStatus = "pending"
	ComplianceStatusVerified  ComplianceStatus = "verified"
	ComplianceStatusRejected  ComplianceStatus = "rejected"
	ComplianceStatusSuspended ComplianceStatus = "suspended"
)

// Dispensary represents a cannabis dispensary
type Dispensary struct {
	BaseModel
	Name              string `json:"name" gorm:"not null;index"`
	LicenseNumber     string `json:"license_number" gorm:"not null;unique;index"`
	State             string `json:"state" gorm:"not null;size:2;index"`
	City              string `json:"city" gorm:"not null;index"`
	Address           string `json:"address" gorm:"not null"`
	Phone             string `json:"phone" gorm:"not null"`
	Email             string `json:"email" gorm:"not null;index"`
	Website           string `json:"website"`
	Status            string `json:"status" gorm:"not null;default:'active';index"`
	ComplianceStatus  ComplianceStatus `json:"compliance_status" gorm:"not null;default:'pending';index"`
	
	// Cannabis-specific fields
	OperatingHours    string `json:"operating_hours"`
	DeliveryAvailable bool   `json:"delivery_available" gorm:"default:false"`
	PickupAvailable   bool   `json:"pickup_available" gorm:"default:true"`
	
	// Compliance fields
	LastInspectionDate   *time.Time `json:"last_inspection_date"`
	NextInspectionDate   *time.Time `json:"next_inspection_date"`
	ComplianceNotes      string     `json:"compliance_notes"`
	
	// Relationships
	Users        []User        `json:"users" gorm:"foreignKey:DispensaryID"`
	Products     []Product     `json:"products" gorm:"foreignKey:DispensaryID"`
	Orders       []Order       `json:"orders" gorm:"foreignKey:DispensaryID"`
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:DispensaryID"`
}

// User represents a cannabis platform user
type User struct {
	BaseModel
	DispensaryID uuid.UUID `json:"dispensary_id" gorm:"type:uuid;not null;index"`
	
	// Basic user information
	FirstName string `json:"first_name" gorm:"not null"`
	LastName  string `json:"last_name" gorm:"not null"`
	Email     string `json:"email" gorm:"not null;unique;index"`
	Phone     string `json:"phone" gorm:"index"`
	
	// Authentication fields
	PasswordHash string `json:"-" gorm:"not null"`
	IsActive     bool   `json:"is_active" gorm:"default:true;index"`
	Role         UserRole `json:"role" gorm:"not null;index"`
	
	// Cannabis compliance fields
	DateOfBirth          *time.Time       `json:"date_of_birth"`
	AgeVerified          bool             `json:"age_verified" gorm:"default:false;index"`
	AgeVerificationMethod string          `json:"age_verification_method"`
	AgeVerifiedAt        *time.Time       `json:"age_verified_at"`
	State                string           `json:"state" gorm:"size:2;index"`
	StateVerified        bool             `json:"state_verified" gorm:"default:false;index"`
	StateVerifiedAt      *time.Time       `json:"state_verified_at"`
	ComplianceStatus     ComplianceStatus `json:"compliance_status" gorm:"not null;default:'pending';index"`
	
	// Address information
	Address     string `json:"address"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
	
	// Profile fields
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Preferences string `json:"preferences" gorm:"type:jsonb"`
	
	// Cannabis-specific preferences
	PreferredProducts    string `json:"preferred_products" gorm:"type:jsonb"`
	ConsumptionGoals     string `json:"consumption_goals" gorm:"type:jsonb"`
	MedicalRecommendation bool  `json:"medical_recommendation" gorm:"default:false"`
	
	// Timestamps
	LastLoginAt           *time.Time `json:"last_login_at"`
	LastComplianceCheckAt *time.Time `json:"last_compliance_check_at"`
	
	// Relationships
	Dispensary   Dispensary    `json:"dispensary" gorm:"foreignKey:DispensaryID"`
	Sessions     []Session     `json:"sessions" gorm:"foreignKey:UserID"`
	Orders       []Order       `json:"orders" gorm:"foreignKey:UserID"`
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:UserID"`
}

// Session represents a user session (Redis-backed)
type Session struct {
	BaseModel
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	DispensaryID  uuid.UUID `json:"dispensary_id" gorm:"type:uuid;not null;index"`
	
	// Session fields
	SessionID     string    `json:"session_id" gorm:"not null;unique;index"`
	AccessToken   string    `json:"access_token" gorm:"not null;unique"`
	RefreshToken  string    `json:"refresh_token" gorm:"not null;unique"`
	ExpiresAt     time.Time `json:"expires_at" gorm:"not null;index"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at" gorm:"not null;index"`
	
	// Session metadata
	IPAddress     string `json:"ip_address"`
	UserAgent     string `json:"user_agent"`
	DeviceType    string `json:"device_type"`
	Location      string `json:"location"`
	
	// Cannabis compliance context
	ComplianceVerified bool       `json:"compliance_verified" gorm:"default:false"`
	ComplianceCheckedAt *time.Time `json:"compliance_checked_at"`
	
	// Session status
	IsActive      bool       `json:"is_active" gorm:"default:true;index"`
	LastActivity  time.Time  `json:"last_activity" gorm:"not null"`
	
	// Relationships
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Dispensary Dispensary `json:"dispensary" gorm:"foreignKey:DispensaryID"`
}

// Product represents a cannabis product
type Product struct {
	BaseModel
	DispensaryID uuid.UUID `json:"dispensary_id" gorm:"type:uuid;not null;index"`
	
	// Product information
	Name        string `json:"name" gorm:"not null;index"`
	Description string `json:"description"`
	SKU         string `json:"sku" gorm:"not null;unique;index"`
	Brand       string `json:"brand" gorm:"not null;index"`
	Category    string `json:"category" gorm:"not null;index"`
	Subcategory string `json:"subcategory" gorm:"index"`
	
	// Cannabis-specific fields
	THCPercentage float64 `json:"thc_percentage" gorm:"index"`
	CBDPercentage float64 `json:"cbd_percentage" gorm:"index"`
	StrainType    string  `json:"strain_type" gorm:"index"` // indica, sativa, hybrid
	
	// Pricing and inventory
	Price        float64 `json:"price" gorm:"not null"`
	Weight       float64 `json:"weight"` // in grams
	Quantity     int     `json:"quantity" gorm:"not null;default:0"`
	Unit         string  `json:"unit" gorm:"not null"`
	
	// Product status
	IsActive     bool   `json:"is_active" gorm:"default:true;index"`
	IsAvailable  bool   `json:"is_available" gorm:"default:true;index"`
	IsFeatured   bool   `json:"is_featured" gorm:"default:false;index"`
	
	// Compliance fields
	LabTested    bool   `json:"lab_tested" gorm:"default:false"`
	LabResults   string `json:"lab_results" gorm:"type:jsonb"`
	ComplianceID string `json:"compliance_id" gorm:"index"`
	
	// Media
	Images       string `json:"images" gorm:"type:jsonb"`
	
	// Relationships
	Dispensary   Dispensary `json:"dispensary" gorm:"foreignKey:DispensaryID"`
	OrderItems   []OrderItem `json:"order_items" gorm:"foreignKey:ProductID"`
}

// Order represents a cannabis order
type Order struct {
	BaseModel
	DispensaryID uuid.UUID `json:"dispensary_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	
	// Order information
	OrderNumber string  `json:"order_number" gorm:"not null;unique;index"`
	Status      string  `json:"status" gorm:"not null;index"`
	Type        string  `json:"type" gorm:"not null;index"` // pickup, delivery
	
	// Pricing
	Subtotal    float64 `json:"subtotal" gorm:"not null"`
	Tax         float64 `json:"tax" gorm:"not null"`
	Discount    float64 `json:"discount" gorm:"default:0"`
	Total       float64 `json:"total" gorm:"not null"`
	
	// Cannabis compliance fields
	AgeVerified          bool      `json:"age_verified" gorm:"not null;default:false"`
	StateVerified        bool      `json:"state_verified" gorm:"not null;default:false"`
	ComplianceVerifiedAt time.Time `json:"compliance_verified_at"`
	
	// Order details
	DeliveryAddress string     `json:"delivery_address"`
	DeliveryNotes   string     `json:"delivery_notes"`
	ScheduledFor    *time.Time `json:"scheduled_for"`
	CompletedAt     *time.Time `json:"completed_at"`
	
	// Relationships
	Dispensary Dispensary  `json:"dispensary" gorm:"foreignKey:DispensaryID"`
	User       User        `json:"user" gorm:"foreignKey:UserID"`
	Items      []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

// OrderItem represents an item in a cannabis order
type OrderItem struct {
	BaseModel
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	
	// Item details
	Quantity    int     `json:"quantity" gorm:"not null"`
	UnitPrice   float64 `json:"unit_price" gorm:"not null"`
	TotalPrice  float64 `json:"total_price" gorm:"not null"`
	
	// Cannabis compliance tracking
	WeightSold  float64 `json:"weight_sold"` // in grams
	
	// Relationships
	Order   Order   `json:"order" gorm:"foreignKey:OrderID"`
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}

// Transaction represents a payment transaction
type Transaction struct {
	BaseModel
	DispensaryID uuid.UUID `json:"dispensary_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderID      uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	
	// Transaction details
	TransactionID string  `json:"transaction_id" gorm:"not null;unique;index"`
	Amount        float64 `json:"amount" gorm:"not null"`
	Fee           float64 `json:"fee" gorm:"default:0"`
	Status        string  `json:"status" gorm:"not null;index"`
	Method        string  `json:"method" gorm:"not null;index"`
	
	// Payment processor fields
	ProcessorID       string `json:"processor_id"`
	ProcessorResponse string `json:"processor_response" gorm:"type:jsonb"`
	
	// Cannabis compliance audit
	ComplianceLogged bool      `json:"compliance_logged" gorm:"default:false"`
	AuditTrail       string    `json:"audit_trail" gorm:"type:jsonb"`
	
	// Relationships
	Dispensary Dispensary `json:"dispensary" gorm:"foreignKey:DispensaryID"`
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Order      Order      `json:"order" gorm:"foreignKey:OrderID"`
}

// AuditLog represents cannabis compliance audit logs
type AuditLog struct {
	BaseModel
	DispensaryID uuid.UUID `json:"dispensary_id" gorm:"type:uuid;index"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	
	// Audit fields
	Action       string `json:"action" gorm:"not null;index"`
	Resource     string `json:"resource" gorm:"not null;index"`
	ResourceID   string `json:"resource_id" gorm:"index"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	
	// Cannabis compliance context
	ComplianceType string `json:"compliance_type" gorm:"not null;index"`
	AuditData      string `json:"audit_data" gorm:"type:jsonb"`
	
	// Regulatory fields
	Severity       string `json:"severity" gorm:"not null;index"`
	ComplianceRule string `json:"compliance_rule"`
	
	// Relationships (optional)
	Dispensary *Dispensary `json:"dispensary,omitempty" gorm:"foreignKey:DispensaryID"`
	User       *User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// GetAllModels returns all models for auto-migration
func GetAllModels() []interface{} {
	return []interface{}{
		&Dispensary{},
		&User{},
		&Session{},
		&Product{},
		&Order{},
		&OrderItem{},
		&Transaction{},
		&AuditLog{},
	}
}

// Cannabis-specific validation methods

// IsValidAge checks if user meets minimum age requirement
func (u *User) IsValidAge() bool {
	if u.DateOfBirth == nil {
		return false
	}
	
	age := time.Since(*u.DateOfBirth).Hours() / 24 / 365.25
	return age >= 21
}

// IsCompliant checks if user is compliant for cannabis purchases
func (u *User) IsCompliant() bool {
	return u.AgeVerified && u.StateVerified && u.ComplianceStatus == ComplianceStatusVerified
}

// IsValidCannabisState checks if state allows cannabis
func (u *User) IsValidCannabisState() bool {
	legalStates := map[string]bool{
		"CA": true, "CO": true, "WA": true, "OR": true, "NV": true,
		"AZ": true, "NY": true, "IL": true, "NJ": true, "VA": true,
		"CT": true, "MT": true, "VT": true, "AK": true, "MA": true,
		"ME": true, "MI": true, "MD": true, "MO": true, "OH": true,
	}
	
	return legalStates[u.State]
}

// BeforeCreate hook for User model
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for all models with BaseModel
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}