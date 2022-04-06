package main

import (
	"github.com/dimaglushkov/go-chat/app"
	"log"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
