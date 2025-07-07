// Developer: zeelrupapara@gmail.com
// Description: API routes for GreenLync Event-Driven API Gateway Boilerplate

package v1

import (
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/swagger"
	"github.com/gofiber/websocket/v2"
)

func (s *HttpServer) RegisterV1() {
	//************************ Route Grouping *******************************
	root := s.App.Group("/")
	api := root.Group("/api")
	ws := root.Group("/ws/v1")
	oauth := root.Group("/auth/v1/oauth2")
	v1 := api.Group("/v1")
	system := v1.Group("/system")
	_ = v1.Group("/public")

	//************************ Global Middlewares *******************************
	root.Use(cors.New())

	api.Use(s.Middleware.UserAgentParser, s.Middleware.HeaderReader, s.Middleware.RequestsLogger)
	ws.Use(s.Middleware.UserAgentParser, s.Middleware.HeaderReader, s.Middleware.RequestsLogger)
	oauth.Use(s.Middleware.UserAgentParser, s.Middleware.HeaderReader, s.Middleware.RequestsLogger)

	ws.Use(s.Middleware.Protect)
	system.Use(s.Middleware.Protect)

	//************************ AUTH Routes *******************************
	oauth.Post("/login", s.Middleware.BasicAuthParser, s.Login)
	oauth.Post("/token", s.Middleware.BasicAuthParser, s.Token)
	oauth.Post("/refresh/token", s.RefreshToken)
	oauth.Delete("/logout", s.Middleware.Protect, s.Logout)

	//************************ Websocket *****************************
	ws.Get("/", websocket.New(s.serveWS))

	//************************ System Routes *****************************
	monitorRoutes := system.Group("/monitor")
	resourceRoutes := system.Group("/resources")
	policyRoutes := system.Group("/policies")
	tokenRoutes := system.Group("/tokens")
	sessionRoutes := system.Group("/sessions")
	operationRoutes := system.Group("/operations")

	// monitor
	monitorRoutes.Get("/health", s.CheckSystemHealth)

	// Roles
	roleRoutes := system.Group("/roles")
	roleRoutes.Get("/", s.Middleware.Authorization(authz.Resources_Roles_Read), s.GetAllRoles)
	roleRoutes.Post("/", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.CreateRole)
	roleRoutes.Patch("/:role_id", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.UpdateRole)
	roleRoutes.Delete("/:role_id", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.DeleteRole)

	// Resources
	resourceRoutes.Get("/", s.Middleware.Authorization(authz.Resources_Roles_Read), s.GetAllResources)
	resourceRoutes.Get("/:role_id/role", s.Middleware.Authorization(authz.Resources_Roles_Read), s.GetAllResourcesWithRole)

	// Policies
	policyRoutes.Post("/:role_id", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.CreatePolicies)
	policyRoutes.Put("/:role_id", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.UpdatePolicies)
	policyRoutes.Post("/", s.Middleware.Authorization(authz.Resources_Roles_Manage), s.DeletePolicies)

	// Tokens
	tokenRoutes.Get("/history", s.Middleware.Authorization(authz.Resources_Tokens_Read), s.GetAllTokensHistroy)
	tokenRoutes.Delete("/history", s.Middleware.Authorization(authz.Resources_Tokens_Delete), s.DeleteAllTokensHistroy)

	// Sessions
	sessionRoutes.Get("/active", s.Middleware.Authorization(authz.Resources_Sessions_Read), s.GetAllSessions)
	sessionRoutes.Get("/history", s.Middleware.Authorization(authz.Resources_Sessions_Read), s.GetAllSessionsHistroy)
	sessionRoutes.Delete("/history", s.Middleware.Authorization(authz.Resources_Sessions_Delete), s.DeleteAllSessionsHistroy)
	sessionRoutes.Delete("/active", s.Middleware.Authorization(authz.Resources_Sessions_Delete), s.DeleteAllSessions)
	sessionRoutes.Delete("/active/:id", s.Middleware.Authorization(authz.Resources_Sessions_Delete), s.DeleteSession)

	// Operations
	operationRoutes.Get("/", s.Middleware.Authorization(authz.Resources_Logs_Read), s.GetAllOperations)
	operationRoutes.Get("/:operation_id", s.Middleware.Authorization(authz.Resources_Logs_Read), s.GetOperation)
	// this route is against the regulations
	// operationRoutes.Delete("/:operation_id", s.Middleware.Authorization(authz.Resources_OperationLogs_Delete), s.DeleteOperation)
	// operationRoutes.Delete("/", s.Middleware.Authorization(authz.Resources_OperationLogs_Delete), s.DeleteAllOperations)

	//************************ Business Routes *****************************

	// Core business functionality routes
	configRoutes := v1.Group("/configs")
	emailRoutes := v1.Group("/emails")

	// System Configs
	configRoutes.Use(s.Middleware.Protect)
	configRoutes.Get("/", s.Middleware.Authorization(authz.Resources_Config_Read), s.GetAllConfigs)
	configRoutes.Get("/:config_id", s.Middleware.Authorization(authz.Resources_Config_Read), s.GetConfig)
	configRoutes.Patch("/:config_id", s.Middleware.Authorization(authz.Resources_Config_Update), s.UpdateConfig)
	configRoutes.Get("/groups/:group_id", s.Middleware.Authorization(authz.Resources_Config_Read), s.GetConfigsBelongToGroup)


	// InEmails
	emailRoutes.Use(s.Middleware.Protect)
	emailRoutes.Get("/", s.Middleware.Authorization(authz.Resources_Emails_Read), s.GetAllInEmails)
	emailRoutes.Get("/:id", s.Middleware.Authorization(authz.Resources_Emails_Read), s.GetInEmail)
	emailRoutes.Post("/", s.Middleware.Authorization(authz.Resources_MyEmails_Create), s.CreateInEmail)
	emailRoutes.Put("/:id", s.Middleware.Authorization(authz.Resources_MyEmails_Update), s.UpdateInEmail)
	emailRoutes.Delete("/:id", s.Middleware.Authorization(authz.Resources_MyEmails_Delete), s.DeleteInEmail)
	emailRoutes.Get("/me/inbox", s.Middleware.Authorization(authz.Resources_MyEmails_Read), s.GetAllAccountInEmails)
	emailRoutes.Get("/me/outbox", s.Middleware.Authorization(authz.Resources_MyEmails_Read), s.GetAllAccountOutEmails)
	emailRoutes.Get("/me/draft", s.Middleware.Authorization(authz.Resources_MyEmails_Read), s.GetAccountDraftEmails)
	emailRoutes.Get("/me/bin", s.Middleware.Authorization(authz.Resources_MyEmails_Read), s.GetAccountBinEmails)
	emailRoutes.Get("/me/:tracking_id", s.Middleware.Authorization(authz.Resources_MyEmails_Read), s.GetAccountInEmail)


	// in case no API route was found
	api.All("*", func(c *fiber.Ctx) error {
		s.Log.Logger.Error(errors.EndpointNotFound)
		return s.App.HttpResponseNotFound(c, errors.ErrEndpointNotFound)
	})

	//************************ Static Routes *****************************
	// Favicon
	root.Use(favicon.New(favicon.Config{
		File: "./public/static/manifest/favicon.ico",
		URL:  "/static/manifest/favicon.ico",
	}))

	// Swagger - Public access for development
	root.Get("/swagger/*", swagger.HandlerDefault)

	// API Docs
	apiDocs := root.Group("/api/docs")
	apiDocs.Use(s.Middleware.ProtectStatic)
	apiDocs.Static("/", "public/docs/api")

	// Admin Docs
	adminDocs := root.Group("/admin/docs")
	adminDocs.Use(s.Middleware.ProtectStatic, s.Middleware.Authorization(authz.Resources_Type_Admin))
	adminDocs.Static("/", "public/docs/admin")

	// Any other static files under static folder
	root.Static("/static", "public/static")
	root.Static("/others", "public/others")

	// Web
	// get html, css, js and images etc.....
	root.Static("/", "public/web")
	// in case the URL was meant for the web application
	root.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("public/web/index.html")
	})

	root.All("*", func(c *fiber.Ctx) error {
		s.Log.Logger.Error(errors.EndpointNotFound)
		return s.App.HttpResponseNotFound(c, errors.ErrEndpointNotFound)
	})
}
