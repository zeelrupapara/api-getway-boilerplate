// Developer: zeelrupapara@gmail.com
// Description: RBAC and authorization for GreenLync boilerplate

package v1

import (
	"fmt"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CrtAction struct {
	Action  string `json:"action,omitempty"`
	Checked bool   `json:"checked"`
}

type Policy struct {
	Role     string      `json:"role,omitempty"`
	Resource string      `json:"resource,omitempty"`
	Actions  []CrtAction `json:"actions,omitempty"`
}

// ****************************************************************************
// ************************ Roles *********************************************
// ****************************************************************************

// @Id				GetAllRoles
// @Description	Get All System Roles
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Role
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/roles [get]
func (s *HttpServer) GetAllRoles(c *fiber.Ctx) error {
	roles := []model.Role{}
	err := s.DB.Find(&roles).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, &roles)
}

type CrtRole struct {
	Desc     string         `json:"desc" validate:"required" example:"trader"`
	RoleType model.RoleType `json:"role_type"`
}

// @Id				CreateRole
// @Description	Create Role
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		201	{object}	model.Role
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			body	body	CrtRole	true	"Role Request Body"
// @Router			/api/v1/system/roles [post]
// @Hidden
func (s *HttpServer) CreateRole(c *fiber.Ctx) error {
	data := &CrtRole{}
	err := c.BodyParser(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	role := &model.Role{
		Desc:     data.Desc,
		RoleType: data.RoleType,
	}
	err = s.DB.Create(role).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseCreated(c, role)
}

type UptRole struct {
	Desc string `json:"desc" validate:"required"`
}

// @Id				UpdateRole
// @Description	Update Role
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{object}	model.Role
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			role_id	path	int		true	"Role ID"
// @Param			body	body	CrtRole	true	"Role Request Body"
// @Router			/api/v1/system/roles/{role_id} [patch]
func (s *HttpServer) UpdateRole(c *fiber.Ctx) error {
	roleId, err := c.ParamsInt("role_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	role := &model.Role{}
	err = s.DB.First(role, roleId).Error
	if err == gorm.ErrRecordNotFound {
		return s.App.HttpResponseNotFound(c, err)
	} else if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if role.Original {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("cannot update original role"))
	}

	data := &UptRole{}
	err = c.BodyParser(data)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	role.Desc = data.Desc

	tx := s.DB.Begin()
	err = tx.Save(&role).Error
	if err != nil {
		tx.Rollback()
		return s.App.HttpResponseBadRequest(c, err)
	}

	if role.Desc != data.Desc {
		rules := s.Authz.Enforcer.GetFilteredNamedPolicy("p", 0, role.Desc)

		for _, rule := range rules {
			if rule[0] == data.Desc {
				rule[0] = role.Desc // Replace "admin" with "newadmin"
			}
		}

		_, err = s.Authz.Enforcer.RemoveFilteredNamedPolicy("p", 0, role.Desc)
		if err != nil {
			tx.Rollback()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		if len(rules) > 0 {
			_, err = s.Authz.Enforcer.AddNamedPolicies("p", rules)
			if err != nil {
				tx.Rollback()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}

			err = s.saveChanges()
			if err != nil {
				tx.Rollback()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
		}
	}

	tx.Commit()
	return s.App.HttpResponseOK(c, role)
}

// @Id				DeleteRole
// @Description	Delete Role
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		204
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			role_id	path	int	true	"Role ID"
// @Router			/api/v1/system/roles/{role_id} [DELETE]
func (s *HttpServer) DeleteRole(c *fiber.Ctx) error {
	roleId, err := c.ParamsInt("role_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	role := &model.Role{}
	err = s.DB.First(role, roleId).Error
	if err == gorm.ErrRecordNotFound {
		return s.App.HttpResponseNotFound(c, err)
	} else if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if role.Original {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("cannot update original role"))
	}

	account := &model.User{}
	err = s.DB.Where("role = ?", role.Desc).First(account).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if account.Id != 0 {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("can't delete role with associated account"))
	}

	tx := s.DB.Begin()
	res := s.DB.Delete(role, roleId)
	if res.Error != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// delete all associated policies related to this role
	// _, err = s.Authz.Enforcer.RemoveFilteredNamedPolicy("p", 0, strconv.FormatInt(int64(role.RoleId), 10))
	_, err = s.Authz.Enforcer.RemoveFilteredNamedPolicy("p", 0, role.Desc)
	if err != nil {
		tx.Rollback()
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	err = s.saveChanges()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	tx.Commit()
	return s.App.HttpResponseNoContent(c)
}

// ****************************************************************************
// ************************ Resources *****************************************
// ****************************************************************************

// @Id				GetAllResources
// @Description	Get All Resources Defined in the system
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Resource
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/resources [get]
func (s *HttpServer) GetAllResources(c *fiber.Ctx) error {
	resources := []model.Resource{}
	err := s.DB.Preload("Actions").Find(&resources).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, &resources)
}

// @Id				GetAllResourcesWithRole
// @Description	Get All Resources defined in the system with role permission on them
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Resource
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			role_id	path	int	true	"Role ID"
// @Router			/api/v1/system/resources/{role_id}/role [get]
func (s *HttpServer) GetAllResourcesWithRole(c *fiber.Ctx) error {
	roleId, err := c.ParamsInt("role_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	role := &model.Role{}
	err = s.DB.First(role, roleId).Error
	if err == gorm.ErrRecordNotFound {
		return s.App.HttpResponseBadRequest(c, err)
	} else if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rulesStr := s.Authz.Enforcer.GetFilteredNamedPolicy("p", 0, role.Desc)

	// resources map point to array of actions
	rm := roleResourcesMap(rulesStr)

	resources := []model.Resource{}
	err = s.DB.Preload("Actions").Find(&resources).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	for _, r := range resources {
		roleActions := rm[r.Desc]
		for j, a := range r.Actions {
			for _, pa := range roleActions {
				if pa.Action == a.Desc {
					r.Actions[j].Checked = true
				}
			}
		}
	}

	return s.App.HttpResponseOK(c, &resources)
}

// ****************************************************************************
// ************************ Policies ******************************************
// ****************************************************************************

// @Id				GetMyPolicies
// @Description	Get My Policies in a Map belongs to a Role For UI usage
// @Tags			Accounts
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]bool
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/accounts/me/policies/ui [get]
func (s *HttpServer) GetMyPolicies(c *fiber.Ctx) error {
	client, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
	}

	role := &model.Role{}
	err := s.DB.Where("`desc` = ?", client.Scope).First(role).Error
	if err == gorm.ErrRecordNotFound {
		return s.App.HttpResponseBadRequest(c, err)
	} else if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rulesStr := s.Authz.Enforcer.GetFilteredNamedPolicy("p", 0, role.Desc)
	// resources map point to array of actions
	rm := roleResourcesMap(rulesStr)
	resources := []model.Resource{}
	err = s.DB.Preload("Actions").Find(&resources).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	policiesMap := make(map[string]bool)
	for _, r := range resources {
		roleActions := rm[r.Desc]
		for _, a := range r.Actions {
			// policiesMap[fmt.Sprintf("%s_%s", r.Desc, a.Desc)] = false
			for _, pa := range roleActions {
				if pa.Action == a.Desc {
					policiesMap[fmt.Sprintf("%s_%s", r.Desc, a.Desc)] = true
					break
				}
			}
		}
	}

	return s.App.HttpResponseOK(c, policiesMap)
}

// @Id				CreatePolicies
// @Description	Create Policies
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		201	{array}		Policy
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			role_id	path	int			true	"Role ID"
// @Param			body	body	[]Policy	true	"Create Policies Request body"
// @Router			/api/v1/system/policies/{role_id} [post]
func (s *HttpServer) CreatePolicies(c *fiber.Ctx) error {
	roleId, err := c.ParamsInt("role_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	role := &model.Role{}
	err = s.DB.First(role, roleId).Error
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	policies := []Policy{}
	err = c.BodyParser(&policies)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("error parsing policy struct %v", err))
	}
	rulesStr := [][]string{}
	for i := range policies {
		// assign Role
		policies[i].Role = role.Desc

		// check resource exists
		resource := &model.Resource{}
		err = s.DB.Preload("Actions").First(resource, &model.Resource{Desc: policies[i].Resource}).Error
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, fmt.Errorf("resource doesn't exist"))
		} else if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		for j := range policies[i].Actions {
			action := policies[i].Actions[j].Action
			if err != nil {
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
			doesActionExist := false
			for _, a := range resource.Actions {
				if a.Desc == action {
					doesActionExist = true
					break
				}
			}
			if !doesActionExist {
				return s.App.HttpResponseNotFound(c, fmt.Errorf("action doesn't exist"))
			}
		}

		rulesStr = append(rulesStr, policyToString(&policies[i])...)
	}

	_, err = s.Authz.Enforcer.AddNamedPolicies("p", rulesStr)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	err = s.saveChanges()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseCreated(c, &policies)
}

// @Id				UpdatePolicies
// @Description	Update Policies
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	boolean		boolean
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			role_id	path	int			true	"Role ID"
// @Param			body	body	[]Policy	true	"Update Policies Request body"
// @Router			/api/v1/system/policies/{role_id} [put]
func (s *HttpServer) UpdatePolicies(c *fiber.Ctx) error {
	roleId, err := c.ParamsInt("role_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	role := &model.Role{}
	err = s.DB.First(role, roleId).Error
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	policies := []Policy{}
	err = c.BodyParser(&policies)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	_, err = s.Authz.Enforcer.RemoveFilteredNamedPolicy("p", 0, role.Desc)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	newRules := [][]string{}
	for i := range policies {
		// assign Role
		policies[i].Role = role.Desc

		// check resources, roles and actions exists
		resource := &model.Resource{}
		err = s.DB.Preload("Actions").First(resource, &model.Resource{Desc: policies[i].Resource}).Error
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseNotFound(c, fmt.Errorf("resource doesn't exist"))
		} else if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		for j := range policies[i].Actions {
			action := policies[i].Actions[j].Action
			if err != nil {
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
			doesActionExist := false
			for _, a := range resource.Actions {
				if a.Desc == action {
					doesActionExist = true
					break
				}
			}
			if !doesActionExist {
				return s.App.HttpResponseNotFound(c, fmt.Errorf("action doesn't exist"))
			}
		}

		newRules = append(newRules, policyToString(&policies[i])...)
	}
	updated, err := s.Authz.Enforcer.AddNamedPolicies("p", newRules)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	err = s.saveChanges()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, updated)
}

// @Id				DeletePolicies
// @Description	Delete Policies
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		204
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			body	body	[]Policy	true	"Delete Policies Request body"
// @Router			/api/v1/system/policies [POST]
func (s *HttpServer) DeletePolicies(c *fiber.Ctx) error {
	policies := []Policy{}
	err := c.BodyParser(&policies)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("error parsing policy struct %v", err))
	}

	rulesStr := mapPoliciesToString(policies)
	_, err = s.Authz.Enforcer.RemoveNamedPolicies("p", rulesStr)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	err = s.saveChanges()
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseNoContent(c)
}

// ****************************************************************************
// ************************ Local methods *************************************
// ****************************************************************************
// local methods
func (s *HttpServer) saveChanges() error {
	return s.Authz.Enforcer.SavePolicy()
}

func policyToString(p *Policy) (rules [][]string) {

	for _, a := range p.Actions {
		rules = append(rules, []string{p.Role, p.Resource, a.Action})
	}
	return rules
}

func mapPoliciesToString(policies []Policy) (rules [][]string) {
	for _, p := range policies {
		rules = append(rules, policyToString(&p)...)
	}
	return
}

func roleResourcesMap(rules [][]string) map[string][]CrtAction {
	resourceMap := make(map[string][]CrtAction)
	for _, rr := range rules {
		resource := string(rr[1])
		action := string(rr[2])

		_, ok := resourceMap[resource]
		if !ok {
			resourceMap[resource] = []CrtAction{}

			resourceMap[resource] = append(resourceMap[resource],
				CrtAction{
					Action:  action,
					Checked: true,
				})
		} else {
			resourceMap[resource] = append(resourceMap[resource], CrtAction{Action: action, Checked: true})
		}
	}
	return resourceMap
}
