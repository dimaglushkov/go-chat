package main

import (
	"log"

	"github.com/dimaglushkov/go-chat/internal/chat"
)

func main() {
	application := chat.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
