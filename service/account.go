package service

type AccountService struct {
	*Server
}

func (AccountService) Initialize(server *Server) AccountService {
	return AccountService{
		Server: server,
	}
}
