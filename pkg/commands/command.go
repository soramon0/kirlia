package commands

import (
	"fmt"
	"os"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

func Run(args []string) error {
	if len(args) < 2 {
		usage()
	}

	if args[1] != "index" {
		fmt.Println("Error: invalid command")
		usage()
	}

	if len(args) < 3 || args[2] == "" {
		fmt.Println("Error: file name is required")
		usage()
	}

	targetPath := args[2]
	tfIndex, err := termfreq.NewIndex(targetPath)
	if err != nil {
		return err
	}

	fmt.Printf("Indexed %d files in %s ...\n", len(tfIndex), targetPath)
	return nil
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  kirlia [command] (options)")
	fmt.Println("    - available commands:")
	fmt.Println("      - index filename")
	os.Exit(1)
}
