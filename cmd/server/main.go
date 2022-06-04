package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func main() {

	db, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Could not connect to db: %s\n", err.Error())
	}
	defer db.Close(context.Background())

	r := gin.Default()
	r.Run()

}
