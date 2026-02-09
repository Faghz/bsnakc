package bootstrap

import (
	"log"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
)

func initConfig() (*config.Config, error) {
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
