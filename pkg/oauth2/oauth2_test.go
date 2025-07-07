// Developer: Saif Hamdan
// Date: 18/7/2023

package oauth2

// func TestOAuth2(t *testing.T) {
// 	redis, err := redis.NewRedisClient(&config.Config{
// 		Redis: config.Redis{
// 			RedisAddr:     "0.0.0.0:6379",
// 			RedisPassword: "null",
// 		},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, redis)

// 	cache := cache.NewCache(redis)
// 	require.NotNil(t, cache)

// 	db, err := db.NewMysqDB(&config.Config{
// 		MySQL: config.MySQL{
// 			MysqlHost:     "127.0.0.1",
// 			MysqlPort:     "3306",
// 			MysqlUser:     "vfxuser",
// 			MysqlPassword: "root@12345",
// 			MysqlDBName:   "vfxcore",
// 		},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, db)

// 	oauth2 := NewOAuth2(cache, db.DB, &config.Config{
// 		HTTP: config.Http{
// 			OAuthTokenExpiresIn: 3600,
// 		},
// 	})
// 	require.NotNil(t, oauth2)

// 	tokenStr := oauth2.GenerateToken()
// 	require.NotEmpty(t, tokenStr)

// 	t.Run("InspectToken", func(t *testing.T) {
// 		ctx := context.Background()
// 		token, err := oauth2.PasswordCredentialsToken(ctx, &Config{

// 			ClientName:     "Saif",
// 			ClientId:       1,
// 			ClientSecretId: "1",
// 			IpAddress:      "0.0.0.0",
// 			Email:          "saif@vfx.com",
// 			Username:       "saifhamdan",
// 		})
// 		require.NoError(t, err)
// 		require.NotNil(t, token)

// 		testCases := []struct {
// 			name          string
// 			token         string
// 			checkResponse func(t *testing.T, cfg *Config, err error)
// 		}{
// 			{
// 				name:  "OK",
// 				token: token.AccessToken,
// 				checkResponse: func(t *testing.T, cfg *Config, err error) {
// 					require.NoError(t, err)
// 					require.NotNil(t, token)
// 				},
// 			},
// 			{
// 				name:  "InvalidTokenString",
// 				token: "invalid",
// 				checkResponse: func(t *testing.T, cfg *Config, err error) {
// 					require.Error(t, err)
// 					require.Nil(t, cfg)
// 				},
// 			},
// 		}

// 		for i := range testCases {
// 			t.Run(testCases[i].name, func(t *testing.T) {
// 				cfg, err := oauth2.Inspect(ctx, testCases[i].token)
// 				testCases[i].checkResponse(t, cfg, err)
// 			})
// 		}
// 	})

// 	t.Run("VerifyToken", func(t *testing.T) {
// 		ctx := context.Background()
// 		token, err := oauth2.PasswordCredentialsToken(ctx, &Config{
// 			ClientName:     "Saif",
// 			ClientId:       1,
// 			ClientSecretId: "1",
// 			IpAddress:      "0.0.0.0",
// 			Email:          "saif@vfx.com",
// 			Username:       "saifhamdan",
// 		})
// 		require.NoError(t, err)
// 		require.NotNil(t, token)

// 		testCases := []struct {
// 			name          string
// 			token         string
// 			checkResponse func(t *testing.T, flag bool)
// 		}{
// 			{
// 				name:  "OK",
// 				token: token.AccessToken,
// 				checkResponse: func(t *testing.T, flag bool) {
// 					require.True(t, flag)
// 				},
// 			},
// 			{
// 				name:  "InvalidTokenString",
// 				token: "randomtest",
// 				checkResponse: func(t *testing.T, flag bool) {
// 					require.False(t, flag)
// 				},
// 			},
// 		}

// 		for i := range testCases {
// 			t.Run(testCases[i].name, func(t *testing.T) {
// 				isVerified := oauth2.Verify(ctx, testCases[i].token)
// 				testCases[i].checkResponse(t, isVerified)
// 			})
// 		}
// 	})

// 	t.Run("RefreshToken", func(t *testing.T) {
// 		ctx := context.Background()
// 		token, err := oauth2.PasswordCredentialsToken(ctx, &Config{
// 			ClientName:     "Saif",
// 			ClientId:       1,
// 			ClientSecretId: "1",
// 			IpAddress:      "0.0.0.0",
// 			Email:          "saif@vfx.com",
// 			Username:       "saifhamdan",
// 		})
// 		require.NoError(t, err)
// 		require.NotNil(t, token)

// 		testCases := []struct {
// 			name          string
// 			token         *Config
// 			checkResponse func(t *testing.T, cfg *Config, err error)
// 		}{
// 			{
// 				name: "InvalidTokenString",
// 				token: &Config{
// 					AccessToken: "1323",
// 				},
// 				checkResponse: func(t *testing.T, cfg *Config, err error) {
// 					require.Error(t, err)
// 					require.Nil(t, cfg)
// 				},
// 			},
// 			{
// 				name: "InvalidRefreshTokenString",
// 				token: &Config{
// 					AccessToken:  token.AccessToken,
// 					RefreshToken: "d12d12d",
// 				},
// 				checkResponse: func(t *testing.T, cfg *Config, err error) {
// 					t.Log(err)
// 					require.Error(t, err)
// 					require.Nil(t, token)
// 				},
// 			},
// 			{
// 				name:  "OK",
// 				token: token,
// 				checkResponse: func(t *testing.T, cfg *Config, err error) {
// 					require.NoError(t, err)
// 					require.NotNil(t, token)
// 				},
// 			},
// 		}

// 		for i := range testCases {
// 			t.Run(testCases[i].name, func(t *testing.T) {
// 				newToken, err := oauth2.RefreshToken(ctx, testCases[i].token.AccessToken, testCases[i].token.RefreshToken)
// 				testCases[i].checkResponse(t, newToken, err)
// 			})
// 		}
// 	})
// 	t.Run("DeleteToken", func(t *testing.T) {
// 		ctx := context.Background()
// 		token, err := oauth2.PasswordCredentialsToken(ctx, &Config{
// 			ClientName:     "Saif",
// 			ClientId:       1,
// 			ClientSecretId: "1",
// 			IpAddress:      "0.0.0.0",
// 			Email:          "saif@vfx.com",
// 			Username:       "saifhamdan",
// 		})
// 		require.NoError(t, err)
// 		require.NotNil(t, token)

// 		testCases := []struct {
// 			name          string
// 			token         *Config
// 			checkResponse func(t *testing.T, err error)
// 		}{
// 			{
// 				name:  "OK",
// 				token: token,
// 				checkResponse: func(t *testing.T, err error) {
// 					require.NoError(t, err)
// 				},
// 			},
// 			{
// 				name: "InvalidTokenString",
// 				token: &Config{
// 					AccessToken:  "d",
// 					RefreshToken: "d",
// 				},
// 				checkResponse: func(t *testing.T, err error) {
// 					require.Error(t, err)
// 				},
// 			},
// 		}

// 		for i := range testCases {
// 			t.Run(testCases[i].name, func(t *testing.T) {
// 				err := oauth2.DeleteToken(ctx, testCases[i].token.AccessToken)
// 				testCases[i].checkResponse(t, err)
// 			})
// 		}
// 	})
// }
