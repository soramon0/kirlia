package main

import (
	"log"
	"os"

	"github.com/soramon0/kirlia/pkg/commands"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
