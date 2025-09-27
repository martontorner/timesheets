package cmd

import (
	"cmp"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/tornermarton/timesheets/internal/cli"
)

func getBuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return bi
}

func getKey(bi *debug.BuildInfo, key string) string {
	if bi == nil {
		return ""
	}

	for _, iter := range bi.Settings {
		if iter.Key == key {
			return iter.Value
		}
	}

	return ""
}

func getGitCommit(bi *debug.BuildInfo) string {
	return getKey(bi, "vcs.revision")
}

func getGitTreeState(bi *debug.BuildInfo) string {
	modified := getKey(bi, "vcs.modified")

	if modified == "true" {
		return "dirty"
	}
	if modified == "false" {
		return "clean"
	}

	return ""
}

func getBuildDate(bi *debug.BuildInfo) string {
	return getKey(bi, "vcs.time")
}

func version(context *cli.Context) {
	bi := getBuildInfo()

	fmt.Printf(`
 _   _                     _               _
| |_(_)_ __ ___   ___  ___| |__   ___  ___| |_ ___
| __| | '_ ' _ \ / _ \/ __| '_ \ / _ \/ _ \ __/ __|
| |_| | | | | | |  __/\__ \ | | |  __/  __/ |_\__ \
 \__|_|_| |_| |_|\___||___/_| |_|\___|\___|\__|___/


timesheets: Work log synchronization, with ease.

GitVersion:    %s
GitCommit:     %s
GitTreeState:  %s
BuildDate:     %s
GoVersion:     %s
Compiler:      %s
Platform:      %s/%s

For more information, see https://github.com/tornermarton/timesheets
`,
		cmp.Or(context.Version, "(devel)"),
		cmp.Or(getGitCommit(bi), "unknown"),
		cmp.Or(getGitTreeState(bi), "unknown"),
		cmp.Or(getBuildDate(bi), "unknown"),
		runtime.Version(),
		runtime.Compiler,
		runtime.GOOS, runtime.GOARCH,
	)

}

func Version(args []string, context *cli.Context) {
	command := &cli.FlagSet{FlagSet: flag.NewFlagSet("version", flag.ExitOnError)}

	command.Usage = func() {
		fmt.Printf(`Usage: timesheets version

Print version information about the timesheets CLI.

For more information, visit: https://github.com/tornermarton/timesheets
`)
	}

	command.Parse(args)
	if command.NArg() > 0 {
		command.Usage()
		os.Exit(1)
	}

	version(context)
}
