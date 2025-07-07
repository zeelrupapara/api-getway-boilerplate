package v1

// func TestGetAllSessions(t *testing.T) {
// 	server := NewTestServer(t)

// 	token := server.GetTestToken(t)

// 	url := "/api/v1/system/sessions"
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	require.NoError(t, err)

// 	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))

// 	resp, err := server.App.Test(req)
// 	require.NoError(t, err)

// 	bb := ResponseCheckerOk(t, resp, 200)

// 	sessions := []model.Session{}
// 	err = json.Unmarshal(bb, &sessions)
// 	require.NoError(t, err)

// 	require.NotNil(t, sessions)
// }

// func TestDeleteSession(t *testing.T) {
// 	server := NewTestServer(t)
// 	token := server.GetTestToken(t)

// 	session := &model.Session{}

// 	err := server.StoreSession(session)
// 	require.NoError(t, err)

// 	testCases := []struct {
// 		name          string
// 		route         string
// 		checkResponse func(t *testing.T, resp *http.Response)
// 	}{
// 		{
// 			name: "InvalidSessionid",
// 			checkResponse: func(t *testing.T, resp *http.Response) {
// 				ResponseCheckerBad(t, resp, 500)
// 			},
// 			route: fmt.Sprint("/api/v1/system/sessions/", shortuuid.New()),
// 		},
// 		{
// 			name: "NoSessionId",
// 			checkResponse: func(t *testing.T, resp *http.Response) {
// 				ResponseCheckerBad(t, resp, 404)
// 			},
// 			route: "/api/v1/system/sessions",
// 		},
// 		{
// 			name: "OK",
// 			checkResponse: func(t *testing.T, resp *http.Response) {
// 				ResponseCheckerNoContent(t, resp)

// 			},
// 			route: fmt.Sprint("/api/v1/system/sessions/", session.Id),
// 		},
// 	}

// 	for i := range testCases {

// 		req, err := http.NewRequest(http.MethodDelete, testCases[i].route, nil)
// 		require.NoError(t, err)

// 		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
// 		req.Header.Set("Content-Type", "application/json")

// 		resp, err := server.App.Test(req)
// 		require.NoError(t, err)

// 		testCases[i].checkResponse(t, resp)
// 	}
// }
