package main

import (
	"log"

	"github.com/mihailtudos/gophermart/internal/app"
)

const (
	configPath = "./configs"
)

func main() {
	if err := app.Run(configPath); err != nil {
		log.Fatal(err)
	}
}
