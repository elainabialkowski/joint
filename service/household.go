package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Household struct {
	Id    int
	Title string
}

type HouseholdRepository struct {
	Db *pgxpool.Pool
}

func (HouseholdRepository) Initialize(db *pgxpool.Pool) HouseholdRepository {
	return HouseholdRepository{
		Db: db,
	}
}

func (repo HouseholdRepository) Get(c context.Context, id int) (Household, error) {
	household := Household{}
	stmt := `select id, title 
			from household 
			where id=$1`
	err := repo.Db.QueryRow(c, stmt, id).Scan(&household.Id, &household.Title)
	if err != nil {
		return Household{}, err
	}
	return household, nil
}

func (repo HouseholdRepository) Create(c context.Context, household Household) (int, error) {
	var id int
	stmt := `insert into household(title) values($1) returning id`
	err := repo.Db.QueryRow(c, stmt, household.Title).Scan(&id)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

type HouseholdService struct {
	*Server
	householdRepo HouseholdRepository
}

func (HouseholdService) Initialize(server *Server) HouseholdService {
	return HouseholdService{
		householdRepo: HouseholdRepository{}.Initialize(server.Db),
	}
}

func (srv HouseholdService) Get(c *gin.Context) {
	id := c.GetInt("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "empty uuid"})
		return
	}

	household, err := srv.householdRepo.Get(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"household": household})

}

func (srv HouseholdService) Create(c *gin.Context) {
	household := Household{}
	err := c.BindJSON(&household)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	id, err := srv.householdRepo.Create(c, household)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})

}
