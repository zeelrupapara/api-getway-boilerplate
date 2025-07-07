package v1

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/shortuuid"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

// ****************************************************************************
// ************************ Roles *********************************************
// ****************************************************************************

func TestGetAllRoles(t *testing.T) {
	server := NewTestServer(t)

	token := server.GetTestToken(t)

	url := "/api/v1/system/roles"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))

	resp, err := server.App.Test(req)
	require.NoError(t, err)

	bb := ResponseCheckerOk(t, resp, 200)

	roles := []model.Role{}
	err = json.Unmarshal(bb, &roles)
	require.NoError(t, err)

	require.NotNil(t, roles)
}

func TestCreateRole(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) interface{}
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) interface{} {
				return &model.Role{
					Desc: shortuuid.New(),
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				bb := ResponseCheckerOk(t, resp, 201)

				role := &model.Role{}
				err := json.Unmarshal(bb, role)
				require.NoError(t, err)
			},
		},
	}

	url := "/api/v1/system/roles"
	for i := range testCases {
		body := testCases[i].requestBody(t)

		b, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

func TestUpdateRole(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) interface{}
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) interface{} {
				return &model.Role{
					Desc: shortuuid.New(),
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				bb := ResponseCheckerOk(t, resp, 200)

				role := &model.Role{}
				err := json.Unmarshal(bb, role)
				require.NoError(t, err)
				require.NotEmpty(t, role)
			},
		},
	}

	role := &model.Role{
		Desc: shortuuid.New(),
	}
	err := server.DB.Create(role).Error
	require.NoError(t, err)

	url := fmt.Sprint("/api/v1/system/roles/", role.Id)
	for i := range testCases {
		body := testCases[i].requestBody(t)

		b, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(b))
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

func TestDeleteRole(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	testCases := []struct {
		name          string
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerNoContent(t, resp)
			},
		},
	}

	role := &model.Role{
		Desc: shortuuid.New(),
	}
	err := server.DB.Create(role).Error
	require.NoError(t, err)

	url := fmt.Sprint("/api/v1/system/roles/", role.Id)
	for i := range testCases {

		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

// ****************************************************************************
// ************************ Resources *****************************************
// ****************************************************************************

func TestGetAllResources(t *testing.T) {
	server := NewTestServer(t)

	token := server.GetTestToken(t)

	url := "/api/v1/system/resources"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))

	resp, err := server.App.Test(req)
	require.NoError(t, err)

	bb := ResponseCheckerOk(t, resp, 200)

	resources := []model.Resource{}
	err = json.Unmarshal(bb, &resources)
	require.NoError(t, err)

	require.NotNil(t, resources)
}

func TestGetAllResourcesForRole(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	role := &model.Role{
		Desc: shortuuid.New(),
	}
	err := server.DB.Create(role).Error
	require.NoError(t, err)

	testCases := []struct {
		name          string
		route         string
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			checkResponse: func(t *testing.T, resp *http.Response) {
				data := ResponseCheckerOk(t, resp, 200)

				resources := []model.Resource{}
				err := json.Unmarshal(data, &resources)
				require.NoError(t, err)
				require.NotEmpty(t, resources)
			},
			route: fmt.Sprintf("/api/v1/system/resources/%d/role", role.Id),
		},
		{
			name: "InvalidRoleIdType",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerBad(t, resp, 400)
			},
			route: fmt.Sprintf("/api/v1/system/resources/%s/role", shortuuid.New()),
		},
		{
			name: "InvalidRoleId",
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerBad(t, resp, 400)
			},
			route: fmt.Sprintf("/api/v1/system/resources/%d/role", 76533),
		},
	}

	for i := range testCases {

		req, err := http.NewRequest(http.MethodGet, testCases[i].route, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

// ****************************************************************************
// ************************ Policies ******************************************
// ****************************************************************************

func TestGetAllPolicies(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	testCases := []struct {
		name          string
		route         string
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			checkResponse: func(t *testing.T, resp *http.Response) {
				data := ResponseCheckerOk(t, resp, 200)

				policy := make(map[string]bool)
				err := json.Unmarshal(data, &policy)
				require.NoError(t, err)
				require.NotEmpty(t, policy)
			},
			route: "/api/v1/system/policies/role/ui",
		},
	}

	for i := range testCases {

		req, err := http.NewRequest(http.MethodGet, testCases[i].route, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

func TestCreatePolicies(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	role := &model.Role{
		Desc: shortuuid.New(),
	}

	err := server.DB.Create(role).Error
	require.NoError(t, err)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) interface{}
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) interface{} {
				return []Policy{
					{
						Resource: "users",
						Actions: []CrtAction{
							{
								Action: "getall",
							},
						},
					},
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				bb := ResponseCheckerOk(t, resp, 201)

				policies := []Policy{}
				err := json.Unmarshal(bb, &policies)
				require.NoError(t, err)
			},
		},
	}

	url := fmt.Sprint("/api/v1/system/policies/", role.Id)
	for i := range testCases {
		body := testCases[i].requestBody(t)

		b, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

func TestUpdatePolicies(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	role := &model.Role{
		Desc: shortuuid.New(),
	}

	err := server.DB.Create(role).Error
	require.NoError(t, err)

	pol := []Policy{
		{
			Resource: "users",
			Actions: []CrtAction{
				{
					Action: "getall",
				},
			},
		},
	}

	url := fmt.Sprint("/api/v1/system/policies/", role.Id)

	b, err := json.Marshal(pol)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	_, err = server.App.Test(req)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) interface{}
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) interface{} {
				return []Policy{
					{
						Resource: "users",
						Actions: []CrtAction{
							{
								Action: "get",
							},
						},
					},
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				bb := ResponseCheckerOk(t, resp, 200)

				var res bool
				err := json.Unmarshal(bb, &res)
				require.NoError(t, err)
				require.True(t, res)
			},
		},
	}

	for i := range testCases {
		body := testCases[i].requestBody(t)

		b, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

func TestDeletePolicies(t *testing.T) {
	server := NewTestServer(t)
	token := server.GetTestToken(t)

	role := &model.Role{
		Desc: shortuuid.New(),
	}

	err := server.DB.Create(role).Error
	require.NoError(t, err)

	pol := []Policy{
		{
			Resource: "users",
			Actions: []CrtAction{
				{
					Action: "getall",
				},
			},
		},
	}

	url := fmt.Sprint("/api/v1/system/policies/", role.Id)

	b, err := json.Marshal(pol)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	_, err = server.App.Test(req)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		requestBody   func(t *testing.T) interface{}
		checkResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "OK",
			requestBody: func(t *testing.T) interface{} {
				return pol
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				ResponseCheckerNoContent(t, resp)
			},
		},
	}

	for i := range testCases {
		body := testCases[i].requestBody(t)

		b, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodDelete, "/api/v1/system/policies/", bytes.NewReader(b))
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		require.NoError(t, err)

		testCases[i].checkResponse(t, resp)
	}
}

// ****************************************************************************
// ************************ Groups ********************************************
// ****************************************************************************

// func TestCreateGroups(t *testing.T) {
// 	server := NewTestServer(t)
// 	token := server.GetTestToken(t)

// 	testCases := []struct {
// 		name          string
// 		requestBody   func(t *testing.T) interface{}
// 		checkResponse func(t *testing.T, resp *http.Response)
// 	}{
// 		{
// 			name: "OK",
// 			requestBody: func(t *testing.T) interface{} {
// 				return &model.Group{
// 					Id: 1,
// 					Root:   true,
// 					Desc:   shortuuid.New(),
// 				}
// 			},
// 			checkResponse: func(t *testing.T, resp *http.Response) {
// 				bb := ResponseCheckerOk(t, resp, 201)

// 				group := &model.Group{}
// 				err := json.Unmarshal(bb, &group)
// 				require.NoError(t, err)
// 			},
// 		},
// 	}

// 	url := "/api/v1/system/groups"
// 	for i := range testCases {
// 		body := testCases[i].requestBody(t)

// 		b, err := json.Marshal(body)
// 		require.NoError(t, err)

// 		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
// 		require.NoError(t, err)

// 		req.Header.Set("Authorization", fmt.Sprint("Bearer", " ", token.AccessToken))
// 		req.Header.Set("Content-Type", "application/json")

// 		resp, err := server.App.Test(req)
// 		require.NoError(t, err)

// 		testCases[i].checkResponse(t, resp)
// 	}
// }


