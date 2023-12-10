package main

import (
	"github.com/nessai1/gophermat/internal/gophermart"
	"log"
)

func main() {
	if err := gophermart.Start(); err != nil {
		log.Fatalf("error while listening application: %s", err.Error())
	}
}
