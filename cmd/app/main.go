package main

import (
	"fmt"
	"log"

	"github.com/kaz-as/test-get-items/config"
	"github.com/kaz-as/test-get-items/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("config error: %s", err)
	}

	err = app.Run(cfg)
	if err != nil {
		return fmt.Errorf("application error: %s", err)
	}

	return nil
}
