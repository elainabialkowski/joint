package service

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	Db       *sqlx.DB
	Cache    *redis.Conn
	Sessions *redis.Conn
}

func (srv Server) Run() error {

	var err error
	srv.Db, err = sqlx.Open("pgx", os.Getenv("PG_URI"))
	if err != nil {
		return err
	}

	cache, err := redis.DialURLContext(context.Background(), os.Getenv("CACHE_URI"), redis.DialOption{})
	if err != nil {
		return err
	}
	srv.Cache = &cache

	sessions, err := redis.DialURLContext(context.Background(), os.Getenv("SESSIONS_URI"), redis.DialOption{})
	if err != nil {
		return err
	}
	srv.Sessions = &sessions

	plaidService := PlaidService{}.Initialize(&srv)
	userService := UserService{}.Initialize(&srv)

	router := gin.Default()
	router.POST("/plaid/link/token", plaidService.CreateLinkToken)
	router.GET("/plaid/access/token", plaidService.GetAccessToken)
	router.GET("/user/:id", userService.GetUser)

	router.Run(os.Getenv("PORT"))

	return nil
}
