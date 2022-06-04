package service

type ReportService struct {
	*Server
}

func (ReportService) Initialize(server *Server) ReportService {
	return ReportService{
		Server: server,
	}
}
