// Developer: Saif Hamdan
// Date: 18/7/2023

package db

import (
	"testing"
	"greenlync-api-gateway/config"

	"github.com/stretchr/testify/require"
)

func TestNewMysqDB(t *testing.T) {

	testCases := []struct {
		Name          string
		MysqlUser     string
		MysqlPassword string
		MysqlHost     string
		MysqlPort     string
		MysqlDBName   string
		CheckResponse func(t *testing.T, d *MysqlDB, err error)
	}{
		{
			Name:          "OK",
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "3306",
			MysqlUser:     "vfxuser",
			MysqlPassword: "root@12345",
			MysqlDBName:   "vfxcore",
			CheckResponse: func(t *testing.T, d *MysqlDB, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, d)
			},
		},
		{
			Name:          "InvalidURL",
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "999",
			MysqlUser:     "vfxuser",
			MysqlPassword: "root@12345",
			MysqlDBName:   "vfxcore",
			CheckResponse: func(t *testing.T, d *MysqlDB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
		{
			Name:          "InvalidCredentials",
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "999",
			MysqlUser:     "lulu",
			MysqlPassword: "root@12345",
			MysqlDBName:   "vfxcore",
			CheckResponse: func(t *testing.T, d *MysqlDB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
		{
			Name:          "InvalidDBName",
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "999",
			MysqlUser:     "vfxuser",
			MysqlPassword: "root@12345",
			MysqlDBName:   "haha",
			CheckResponse: func(t *testing.T, d *MysqlDB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].Name, func(t *testing.T) {
			db, err := NewMysqDB(&config.Config{
				MySQL: config.MySQL{
					MysqlHost:     testCases[i].MysqlHost,
					MysqlPort:     testCases[i].MysqlPort,
					MysqlUser:     testCases[i].MysqlUser,
					MysqlPassword: testCases[i].MysqlPassword,
					MysqlDBName:   testCases[i].MysqlDBName,
				},
			})
			testCases[i].CheckResponse(t, db, err)
		})
	}
}

func TestMigrate(t *testing.T) {
	db, err := NewMysqDB(&config.Config{
		MySQL: config.MySQL{
			MysqlHost:     "127.0.0.1",
			MysqlPort:     "3306",
			MysqlUser:     "vfxuser",
			MysqlPassword: "root@12345",
			MysqlDBName:   "vfxcore",
		},
	})
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)
}
