// Developer: zeelrupapara@gmail.com
// Last Update: Cannabis boilerplate conversion
// Update reason: Simplified RBAC resources for boilerplate

package authz

import (
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// Resources for boilerplate authorization
const (
	// Role Types
	Resources_Type_Admin = "type_admin"
	Resources_Type_User  = "type_user"

	// Admin Resources
	Resources_Config_Read   = "config_read"
	Resources_Config_Update = "config_update"
	Resources_Roles_Read    = "roles_read"
	Resources_Roles_Manage  = "roles_manage"
	Resources_Users_Read    = "users_read"
	Resources_Users_Create  = "users_create"
	Resources_Users_Update  = "users_update"
	Resources_Users_Delete  = "users_delete"
	Resources_Sessions_Read = "sessions_read"
	Resources_Sessions_Delete = "sessions_delete"
	Resources_Tokens_Read   = "tokens_read"
	Resources_Tokens_Delete = "tokens_delete"
	Resources_Emails_Read   = "emails_read"
	Resources_Emails_Send   = "emails_send"
	Resources_Logs_Read     = "logs_read"
	Resources_Logs_Delete   = "logs_delete"

	// User Resources
	Resources_MyProfile_Read           = "myprofile_read"
	Resources_MyProfile_Update         = "myprofile_update"
	Resources_MyProfile_ChangePassword = "myprofile_changepassword"
	Resources_MyEmails_Read            = "myemails_read"
	Resources_MyEmails_Create          = "myemails_create"
	Resources_MyEmails_Update          = "myemails_update"
	Resources_MyEmails_Delete          = "myemails_delete"
	Resources_MySessions_Read          = "mysessions_read"
	Resources_MySessions_Delete        = "mysessions_delete"
)

// Default system Roles (can't be changed)
const (
	Roles_Admin = "Admin"
	Roles_User  = "User"
)

type Authz struct {
	DBadapter *gormadapter.Adapter
	Enforcer  *casbin.CachedEnforcer
}

func NewAuthz(db *gorm.DB) (*Authz, error) {
	a, _ := gormadapter.NewAdapterByDB(db)
	e, err := casbin.NewCachedEnforcer("pkg/authz/model.conf", a)
	if err != nil {
		return nil, err
	}

	// Load the policy from DB.
	err = e.LoadPolicy()
	if err != nil {
		return nil, err
	}

	e.SavePolicy()

	return &Authz{
		DBadapter: a,
		Enforcer:  e,
	}, nil
}
