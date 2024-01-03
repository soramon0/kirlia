package main

import (
	"fmt"
	"os"

	"github.com/soramon0/kirlia/pkg/commands"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
