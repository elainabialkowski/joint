package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	*Server
}

type User struct {
	Id        string
	FirstName string
	LastName  string

	HouseholdId uint64
	Accounts    []uint64

	MonthlyIncome float64
}

func (UserService) Initialize(server *Server) UserService {
	return UserService{
		Server: server,
	}
}

func (srv UserService) GetUser(c *gin.Context) {
	id := c.GetString("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "empty uuid",
		})
		return
	}

	user := User{}
	err := srv.Db.GetContext(c, &user, "SELECT * FROM user WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
