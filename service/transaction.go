package service

type TransactionService struct {
	*Server
}

func (TransactionService) Initialize(server *Server) TransactionService {
	return TransactionService{
		Server: server,
	}
}
