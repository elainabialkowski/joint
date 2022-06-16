package main

import (
	"log"

	"github.com/elainabialkowski/joint/service"
)

func main() {
	err := service.Server{}.Run()
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
}
