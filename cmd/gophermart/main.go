package main

import (
	"github.com/nessai1/gophermat/internal/gophermart"
	"log"
)

func main() {
	if err := gophermart.Start(); err != nil {
		log.Fatalf("Error while listening application: %w", err)
	}
}
