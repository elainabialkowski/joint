package service

type GoalService struct {
	*Server
}

func (GoalService) Initialize(server *Server) GoalService {
	return GoalService{
		Server: server,
	}
}
