package main

import (
	"fmt"
	"os"

	"github.com/soramon0/kirlia/pkg/commands"
)

func main() {
	args := []string{"kirlia", "index", "example/file.xhtml"}
	if err := commands.Run(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
