package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	*Server
}

type User struct {
}

func (UserService) Initialize(server *Server) UserService {
	return UserService{
		Server: server,
	}
}

func (srv UserService) GetUser(c *gin.Context) {
	id := c.GetString("uuid")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "empty uuid",
		})
		return
	}

}
