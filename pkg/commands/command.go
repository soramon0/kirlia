package commands

import (
	"fmt"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

func Run(args []string) error {
	if len(args) < 1 {
		usage()
		return fmt.Errorf("error: command is required")
	}

	if args[0] != "index" {
		usage()
		return fmt.Errorf("error: invalid command")
	}

	if len(args) < 2 || args[1] == "" {
		usage()
		return fmt.Errorf("error: file name is required")
	}

	targetPath := args[1]
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
}
