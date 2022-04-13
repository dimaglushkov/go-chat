package main

import (
	"github.com/dimaglushkov/go-chat/chat"
	"log"
)

func main() {
	application := chat.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
