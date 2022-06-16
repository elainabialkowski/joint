package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Server struct {
	*http.Server
	Db       *pgxpool.Pool
	Cache    *redis.Pool
	Sessions *redis.Pool
}

func (srv Server) Run() error {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := srv.setupConnections()
	if err != nil {
		return err
	}

	err = srv.setupEngine()
	if err != nil {
		return err
	}

	err = srv.migrateUp()
	if err != nil {
		return err
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not start listener: %s\n", err)
		} else if err != nil && err == http.ErrServerClosed {
			log.Println("Tearing down resources.")
			srv.migrateDown()
			srv.Db.Close()
			srv.Cache.Close()
			srv.Sessions.Close()
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

func (srv *Server) setupConnections() error {
	var err error
	ticker := time.NewTicker(2 * time.Second)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*20)
	for {
		select {
		case <-ticker.C:
			srv.Db, err = pgxpool.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
		case <-timeout.Done():
			cancel()
			return errors.New("could not connect to postgres")
		}
		if err == nil {
			log.Printf("Successfully connected to db!\n")
			cancel()
			break
		}
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
			return c, nil
		},
	}
	_, err = srv.Sessions.Get().Do("PING")
	if err != nil {
		return err
	}
	log.Printf("Successfully connected to redis!\n")

	return nil
}

func (srv *Server) setupEngine() error {
	plaidService := PlaidService{}.Initialize(srv)

	engine := gin.Default()
	engine.Use(srv.session())
	engine.Use(gin.Recovery())
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
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

	engine.POST("/plaid/link/token", plaidService.CreateLinkToken)
	engine.GET("/plaid/access/token", plaidService.GetAccessToken)

	f, err := os.Create("/server.log")
	if err != nil {
		return err
	}

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	srv.Server = &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: engine,
	}

	return nil
}

func (srv *Server) migrateUp() error {
	p := &postgres.Postgres{}
	d, err := p.Open(os.Getenv("POSTGRES_URI"))
	if err != nil {
		return err
	}
	defer d.Close()

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"dev", d)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (srv *Server) migrateDown() error {
	p := &postgres.Postgres{}
	d, err := p.Open(os.Getenv("POSTGRES_URI"))
	if err != nil {
		return err
	}
	defer d.Close()

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"dev", d)
	if err != nil {
		return err
	}

	err = m.Down()
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) session() gin.HandlerFunc {
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
