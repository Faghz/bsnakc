package main

import (
	"log"

	"github.com/elzestia/go-boilerplate/internal/bootstrap"
)

// @title Go Boilerplate API
// @version 1.0
// @description A scalable HTTP service with hexagonal architecture
// @contact.name API Support
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
