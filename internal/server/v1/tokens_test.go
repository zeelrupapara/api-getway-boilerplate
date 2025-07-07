package v1

import (
	"fmt"
	"net/http"
	"testing"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/shortuuid"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

func TestGetAllTokens(t *testing.T) {
	server := NewTestServer(t)

	token := server.GetTestToken(t)

	url := "/api/v1/system/tokens"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))

	resp, err := server.App.Test(req)
	require.NoError(t, err)

	bb := ResponseCheckerOk(t, resp, 200)

	tokens := []model.Token{}
	err = json.Unmarshal(bb, &tokens)
	require.NoError(t, err)

	require.NotNil(t, tokens)
}

func TestDeleteToken(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	testCases := []struct {
		name          string
		route         string
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "InvalidAccessToken",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerBad(t, resp, 500)
			},
			route: fmt.Sprint("/api/v1/system/tokens?access_token=", shortuuid.New()),
		},
		{
			name: "NoAccessToken",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerBad(t, resp, 400)
			},
			route: "/api/v1/system/tokens",
		},
		{
			name: "OK",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerNoContent(t, resp)
			},
			route: fmt.Sprint("/api/v1/system/tokens?access_token=", token.AccessToken),
		},
	}

	for i := range testCases {
		req, err := http.NewRequest(http.MethodDelete, testCases[i].route, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}
