// Developer: zeelrupapara@gmail.com
// Last Update: Cannabis boilerplate conversion
// Update reason: Simplified RBAC test for boilerplate

package authz

import (
	"testing"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestAuthz(t *testing.T) {
	db, err := db.NewMysqDB(&config.Config{
		MySQL: config.MySQL{
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "3306",
			MysqlUser:     "root",
			MysqlPassword: "password",
			MysqlDBName:   "greenlync",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	authz, err := NewAuthz(db.DB)
	require.NoError(t, err)
	require.NotNil(t, authz)
}
