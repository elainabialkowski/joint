package service

type AuthService struct {
	*Server
}

func (AuthService) Initialize(server *Server) AuthService {
	return AuthService{
		Server: server,
	}
}
