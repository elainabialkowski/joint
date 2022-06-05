package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	*http.Server
	Db       *sqlx.DB
	Cache    *redis.Pool
	Sessions *redis.Pool
	Log      *log.Logger
}

func (srv Server) Run() error {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := srv.Connect()
	if err != nil {
		return err
	}

	plaidService := PlaidService{}.Initialize(&srv)
	userService := UserService{}.Initialize(&srv)

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(srv.GetSession())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.POST("/plaid/link/token", plaidService.CreateLinkToken)
	router.GET("/plaid/access/token", plaidService.GetAccessToken)
	router.GET("/user/:id", userService.GetUser)

	srv.Server = &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not start listener: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}

	log.Println("goodbye! <3")

	return nil
}

func (srv *Server) Connect() error {
	var err error
	srv.Db, err = sqlx.Open("pgx", os.Getenv("PG_URI"))
	if err != nil {
		return err
	}

	srv.Cache = &redis.Pool{
		MaxIdle:     10,
		Wait:        true,
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
		Wait:        true,
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

	return nil
}

func (srv *Server) GetSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.GetHeader("authorization")
		res, err := srv.Sessions.Get().Do("GET", authorization)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"err": err.Error(),
			})
		}

		srv.Sessions.Get().Do("EXPIRE", authorization, 3600)

		ctx.Keys["user_id"] = res.(string)
	}
}
