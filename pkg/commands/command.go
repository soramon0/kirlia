package commands

import (
	"flag"
	"fmt"

	"github.com/soramon0/kirlia/pkg/server"
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

type cmd struct {
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

func (c *cmd) Execute(args []string) error {
	cmdName := c.Name()

	switch cmdName {
	case indexCmd.Name():
		input := c.String("i", "", "input directory or file to index")
		output := c.String("o", "json", "output format: msgpack, json")
		reportSkipped := c.Bool("rs", false, "report skipped file names")
		if err := c.Parse(args); err != nil {
			return err
		}

		args := termfreq.IndexArgs{
			InputFile:     *input,
			OutputFormat:  *output,
			ReportSkipped: *reportSkipped,
		}
		tfIndex, err := termfreq.GenerateIndex(args)
		if err != nil {
			return err
		}
		fmt.Printf("Indexed %d files in %s ...\n", len(tfIndex), *input)

	case serveCmd.Name():
		addr := c.String("h", "0.0.0.0:9001", "host address")
		format := c.String("f", "json", "file format to read: msgpack, json")
		if err := c.Parse(args); err != nil {
			return err
		}

		tfIndex, err := termfreq.ReadIndexFile(*format)
		if err != nil {
			return err
		}
		fmt.Printf("Loaded %d indexed files from index.%s ...\n", len(tfIndex), *format)

		if err := server.NewServer(*addr, &tfIndex).Serve(); err != nil {
			return err
		}

	default:
		return fmt.Errorf("error: unknown %q command", cmdName)
	}

	return nil
}

func parseCommand(args []string) (*cmd, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error: a command is required")
	}

	c := commands[args[0]]
	if c == nil {
		return nil, fmt.Errorf("error: unknown %q command", args[0])
	}

	return &cmd{c}, nil
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  kirlia command [-options]")
	fmt.Println("    - available commands:")
	fmt.Println("      - index")
	fmt.Println("      - serve")
}
