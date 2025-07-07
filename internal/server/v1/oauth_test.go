package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	model "greenlync-api-gateway/model/common/v1"
	app "greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/oauth2"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	server := NewTestServer(t)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T, req *http.Request)
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T, req *http.Request) {
				username := "admin@company.com"
				password := "123"
				req.SetBasicAuth(username, password)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 200, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				require.Equal(t, 200, body.Code)
				require.Equal(t, true, body.Success)
				require.Empty(t, body.Error)

				loginRes := &LoginResponse{}

				bb, err := json.Marshal(body.Data)
				require.NoError(t, err)

				err = json.Unmarshal(bb, loginRes)
				require.NoError(t, err)

				require.NotNil(t, loginRes)
			},
		},
		{
			name: "Invalid User",
			requestBody: func(t *testing.T, req *http.Request) {
				username := "ooooo@company.com"
				password := "123"
				req.SetBasicAuth(username, password)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 401, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				require.Equal(t, 401, body.Code)
				require.Equal(t, false, body.Success)
				require.NotEmpty(t, body.Error)
				require.Empty(t, body.Data)
			},
		},
		{
			name: "Wrong password",
			requestBody: func(t *testing.T, req *http.Request) {
				username := "admin@company.com"
				password := "1234444"
				req.SetBasicAuth(username, password)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 401, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				require.Equal(t, 401, body.Code)
				require.Equal(t, false, body.Success)
				require.NotEmpty(t, body.Error)
				require.Empty(t, body.Data)
			},
		},
	}

	url := "/auth/v1/oauth2/login"
	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, url, nil)
			require.NoError(t, err)
			testCases[i].requestBody(t, request)
			resp, err := server.App.Test(request, -1)
			require.NoError(t, err)
			testCases[i].checkResponse(t, resp)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	server := NewTestServer(t)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) *oauth2.Config
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) *oauth2.Config {
				return server.GetTestToken(t)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 200, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				require.Equal(t, 200, body.Code)
				require.Equal(t, true, body.Success)
				require.Empty(t, body.Error)

				res := &model.Token{}

				bb, err := json.Marshal(body.Data)
				require.NoError(t, err)

				err = json.Unmarshal(bb, res)
				require.NoError(t, err)

				require.NotNil(t, res)
			},
		},
		{
			name: "InvalidRefreshToken",
			requestBody: func(t *testing.T) *oauth2.Config {
				token := server.GetTestToken(t)

				token.RefreshToken = "d12dkdok12"

				return token
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 500, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				// checking erros
				require.Equal(t, 500, body.Code)
				require.Equal(t, false, body.Success)
				require.NotEmpty(t, body.Error)
				require.Empty(t, body.Data)
			},
		},
	}

	url := "/auth/v1/oauth2/refresh/token"
	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			token := testCases[i].requestBody(t)
			body := &RefreshTokenBody{
				RefreshToken: token.RefreshToken,
			}
			b, err := json.Marshal(body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
			require.NoError(t, err)

			request.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
			request.Header.Set("Content-Type", "application/json")

			resp, err := server.App.Test(request, -1)
			require.NoError(t, err)

			testCases[i].checkResponse(t, resp)
		})
	}
}

func TestLogout(t *testing.T) {
	server := NewTestServer(t)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T, req *http.Request)
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T, req *http.Request) {
				token := server.GetTestToken(t)

				// session := &model.Session{
				// 	AccountId: 1,
				// 	IpAddress: "0.0.0.0",
				// }

				// err := server.StoreSession(session)
				// require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 204, resp.StatusCode)
				require.Empty(t, resp.Body)
			},
		},
		{
			name: "MissingAuthHeader",
			requestBody: func(t *testing.T, req *http.Request) {
				server.GetTestToken(t)

				// session := &model.Session{
				// 	AccountId: 1,
				// 	IpAddress: "0.0.0.0",
				// }

				// err := server.StoreSession(session)
				// require.NoError(t, err)
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 401, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				// checking erros
				require.Equal(t, 401, body.Code)
				require.Equal(t, false, body.Success)
				require.NotEmpty(t, body.Error)
				require.Empty(t, body.Data)
			},
		},
		{
			name: "InvalidBearerToken",
			requestBody: func(t *testing.T, req *http.Request) {
				token := server.GetTestToken(t)

				// session := &model.Session{
				// 	AccountId: 1,
				// 	IpAddress: "0.0.0.0",
				// }
				// err := server.StoreSession(session)
				// require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken, "e12e12e12"))
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				require.Equal(t, 404, resp.StatusCode)
				require.NotEmpty(t, resp.Body)

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				body := &app.HttpResponse{}
				err = json.Unmarshal(b, body)
				require.NoError(t, err)

				// checking erros
				require.Equal(t, 404, body.Code)
				require.Equal(t, false, body.Success)
				require.NotEmpty(t, body.Error)
				require.Empty(t, body.Data)
			},
		},
	}

	url := "/auth/v1/oauth2/logout/session"
	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			testCases[i].requestBody(t, request)
			resp, err := server.App.Test(request, -1)
			require.NoError(t, err)

			testCases[i].checkResponse(t, resp)
		})
	}
}
