package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/tornermarton/timesheets/internal/cli"
	cfg "github.com/tornermarton/timesheets/internal/config"
)

func config(context *cli.Context) {
	cfg.Print(context.Config)
}

func Config(args []string, context *cli.Context) {
	command := &cli.FlagSet{FlagSet: flag.NewFlagSet("config", flag.ExitOnError)}

	command.Usage = func() {
		fmt.Printf(`Usage: timesheets config

Print the used configuration.

For more information, visit: https://github.com/tornermarton/timesheets
`)
	}

	command.Parse(args)
	if command.NArg() > 0 {
		command.Usage()
		os.Exit(1)
	}

	config(context)
}
