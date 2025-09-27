package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tornermarton/timesheets/internal/cli"
	"github.com/tornermarton/timesheets/internal/constants"
	"github.com/tornermarton/timesheets/internal/entries"
)

func sync(context *cli.Context, from time.Time, till time.Time, bail bool, dry bool) {
	source, err := entries.NewTimeEntrySource(context.Config.Source)
	if err != nil {
		log.Fatalf("error creating time entry source: %v\n", err)
		return
	}

	target, err := entries.NewTimeEntryTarget(context.Config.Target)
	if err != nil {
		log.Fatalf("error creating time entry target: %v\n", err)
		return
	}

	entries, err := source.PullTimeEntries(from, till)
	if err != nil {
		log.Fatalf("error pulling time entries: %v\n", err)
		return
	}

	for _, entry := range entries {
		fmt.Printf("%s", entry.String())

		if dry {
			fmt.Printf("\n")
			continue
		}

		if err := target.PushTimeEntry(entry); err != nil {
			if bail {
				log.Fatalf("(✘ - %v)\n", err)
			} else {
				fmt.Printf("(✘ - %v)\n", err)
			}
		} else {
			fmt.Printf("(✔)\n")
		}
	}
}

func Sync(args []string, context *cli.Context) {
	command := &cli.FlagSet{FlagSet: flag.NewFlagSet("sync", flag.ExitOnError)}

	fromFlag := command.Time("from", constants.TODAY, "date/datetime to sync work logs from (inclusive)")
	tillFlag := command.Time("till", constants.TOMORROW, "date/datetime to sync work logs till (exclusive)")

	bailFlag := command.Bool("bail", false, "stop the synchronization process on the first error encountered")
	dryFlag := command.Bool("dry", false, "perform a dry run without making any changes")

	command.Usage = func() {
		fmt.Printf(`Usage: timesheets sync [options]

Synchronize your work logs.

Options:

`)
		command.PrintDefaults()
		fmt.Printf(`
Example (synchronize work logs on 2025-06-01 and 2025-06-02):

  timesheets sync --from 2025-06-01 --till 2025-06-03

For more information, visit: https://github.com/tornermarton/timesheets
`)
	}

	command.Parse(args)
	if command.NArg() > 0 {
		command.Usage()
		os.Exit(1)
	}

	sync(context, *fromFlag, *tillFlag, *bailFlag, *dryFlag)
}
