package service

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx"
)

type Server struct {
	Db       *pgx.ConnPool
	Cache    *redis.Conn
	Sessions *redis.Conn
}
