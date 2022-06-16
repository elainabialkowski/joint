package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type User struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
	Household int
}

type UserRepository struct {
	Db *pgxpool.Pool
}

func (UserRepository) Initialize(db *pgxpool.Pool) UserRepository {
	return UserRepository{
		Db: db,
	}
}

func (repo UserRepository) Get(c context.Context, id int) (User, error) {
	user := User{}
	stmt := `select id, first_name, last_name, email, user_household 
			from user 
			where id=$1`
	err := repo.Db.QueryRow(c, stmt, id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.Household)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (repo UserRepository) Create(c context.Context, user User) (int, error) {
	var id int
	stmt := `insert into user(first_name, last_name, email, user_household) values($1, $2, $3, $4) returning id`
	err := repo.Db.QueryRow(c, stmt, user.FirstName, user.LastName, user.Email, user.Household).Scan(&id)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

type UserService struct {
	*Server
	userRepo UserRepository
}

func (UserService) Initialize(server *Server) UserService {
	return UserService{
		userRepo: UserRepository{}.Initialize(server.Db),
	}
}

func (srv UserService) Get(c *gin.Context) {
	id := c.GetInt("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "empty uuid"})
		return
	}

	user, err := srv.userRepo.Get(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})

}

func (srv UserService) Create(c *gin.Context) {
	user := User{}
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	id, err := srv.userRepo.Create(c, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})

}
