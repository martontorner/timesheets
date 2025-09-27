package main

import (
	"flag"
	"fmt"

	"log"
	"os"

	"github.com/tornermarton/timesheets/cmd"
	"github.com/tornermarton/timesheets/internal/cli"
	cfg "github.com/tornermarton/timesheets/internal/config"
)

// Values are set during the build process using -ldflags.
var (
	version string = ""
)

func main() {
	command := &cli.FlagSet{FlagSet: flag.NewFlagSet("timesheets", flag.ExitOnError)}

	configFlag := command.String("config", "~/.config/timesheets/config.yaml", "config path")

	command.Usage = func() {
		fmt.Printf(`Usage: timesheets [options] <command>

Synchronize your work logs between multiple timesheet managing systems with ease.

Commands:

  config    Print the used configuration.
  sync      Synchronize your work logs.
  version   Print version information about the timesheets CLI.

Options:

`)
		command.PrintDefaults()
		fmt.Printf(`
Example (synchronize work logs on 2025-06-01 and 2025-06-02):

  timesheets sync --from 2025-06-01 --till 2025-06-03

For more information, visit: https://github.com/tornermarton/timesheets
`)
	}

	command.Parse(os.Args[1:])

	config, err := cfg.Read(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	context := &cli.Context{
		Version: version,
		Config:  config,
	}

	switch command.Arg(0) {
	case "config":
		cmd.Config(command.Args()[1:], context)
	case "sync":
		cmd.Sync(command.Args()[1:], context)
	case "version":
		cmd.Version(command.Args()[1:], context)
	default:
		command.Usage()
		os.Exit(1)
	}
}
