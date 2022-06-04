package service

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	Db       *sqlx.DB
	Cache    *redis.Pool
	Sessions *redis.Pool
}

func (srv Server) Run() error {

	var err error
	srv.Db, err = sqlx.Open("pgx", os.Getenv("PG_URI"))
	if err != nil {
		return err
	}
	defer srv.Db.Close()

	srv.Cache = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			c, err := redis.Dial("tcp", os.Getenv("CACHE_URI"))
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", os.Getenv("CACHE_PASSWORD")); err != nil {
				c.Close()
				return nil, err
			}
			if _, err := c.Do("SELECT", os.Getenv("CACHE_DB")); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

	srv.Sessions = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			c, err := redis.Dial("tcp", os.Getenv("SESSIONS_URI"))
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", os.Getenv("SESSIONS_PASSWORD")); err != nil {
				c.Close()
				return nil, err
			}
			if _, err := c.Do("SELECT", os.Getenv("SESSIONS_DB")); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

	plaidService := PlaidService{}.Initialize(&srv)
	userService := UserService{}.Initialize(&srv)

	router := gin.Default()

	router.POST("/plaid/link/token", plaidService.CreateLinkToken)
	router.GET("/plaid/access/token", plaidService.GetAccessToken)

	router.GET("/user/:id", userService.GetUser)

	router.Run(os.Getenv("PORT"))

	return nil
}
