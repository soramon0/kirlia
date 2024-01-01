package commands

import (
	"flag"
	"fmt"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

var (
	indexCmd = flag.NewFlagSet("index", flag.ExitOnError)
	serveCmd = flag.NewFlagSet("serve", flag.ExitOnError)
	commands = map[string]*flag.FlagSet{
		indexCmd.Name(): indexCmd,
		serveCmd.Name(): serveCmd,
	}
)

type Cmd struct {
	*flag.FlagSet
}

func Run(args []string) error {
	cmd, err := parseCommand(args)
	if err != nil {
		usage()
		return err
	}

	return cmd.Execute(args[1:])
}

func (c *Cmd) Execute(args []string) error {
	if c == nil {
		return fmt.Errorf("error: uknown command")
	}

	cmdName := c.Name()
	switch cmdName {
	case indexCmd.Name():
		targetPath := c.String("f", "", "directory or file to index")
		reportSkipped := c.Bool("rs", false, "report skipped file names")
		if err := c.Parse(args); err != nil {
			return err
		}
		if *targetPath == "" {
			return fmt.Errorf("error: file name is required")
		}

		tfIndex, err := termfreq.NewIndex(*targetPath, *reportSkipped)
		if err != nil {
			return err
		}
		fmt.Printf("Indexed %d files in %s ...\n", len(tfIndex), *targetPath)

	case serveCmd.Name():
		return fmt.Errorf("error: serve command not implemented yet")

	default:
		return fmt.Errorf("error: unknown %q command", cmdName)
	}

	return nil
}

func parseCommand(args []string) (*Cmd, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error: a command is required")
	}

	cmd := commands[args[0]]
	if cmd == nil {
		return nil, fmt.Errorf("error: uknown %q command", args[0])
	}

	return &Cmd{cmd}, nil
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  kirlia command [-options]")
	fmt.Println("    - available commands:")
	fmt.Println("      - index")
	fmt.Println("      - serve")
}
