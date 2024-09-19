package main

import (
	"log"

	"github.com/mihailtudos/gophermart/internal/app"
)


func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
