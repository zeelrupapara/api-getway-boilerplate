package server

func (s *Server) Register() {
	// OAuth Routes
	// s.OAuth.RegisterOAuth()
	// Register v1 Routes
	s.Web.RegisterV1()
	s.Web.RegisterWSV1()
}
