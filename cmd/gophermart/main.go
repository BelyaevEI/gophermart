package main

import (
	"log"

	"github.com/BelyaevEI/gophermart/internal/app"
)

func main() {

	if err := app.RunServer(); err != nil {
		log.Fatal(err)
	}
}
