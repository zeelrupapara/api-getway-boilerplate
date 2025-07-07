package middleware

import (
	"greenlync-api-gateway/pkg/manager"
)

func (m *Middleware) AuthorizationWS(resource string) manager.Handler {
	// return func(c *manager.Ctx) error {
	// 	r := strings.Split(resource, "_")
	// 	ok := m.Authz.Enforcer.HasNamedPolicy("p", client.Scope, r[0], r[1])
	// 	if !ok {
	// 		return m.App.HttpResponseForbidden(c, errors.ErrUnauthorizedToAccessResource)
	// 	}
	// 	return c.Next()
	// }
	return func(c *manager.Ctx) error { return nil }
}
