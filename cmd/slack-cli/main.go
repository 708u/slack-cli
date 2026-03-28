package main

import (
	"fmt"
	"os"

	// Embed IANA timezone database into the binary so that
	// time.LoadLocation works on minimal environments (e.g.
	// Alpine, scratch containers) that lack /usr/share/zoneinfo.
	_ "time/tzdata"

	"github.com/708u/slack-cli/internal/cmd"
	"github.com/alecthomas/kong"
)

var version = "dev"

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("slack-cli"),
		kong.Description("CLI tool to interact with Slack API"),
		kong.UsageOnError(),
		kong.Vars{"version": version},
	)
	// Resolve timezone before any command runs so that all
	// formatters and parsers share the same *time.Location.
	// Must run after kong.Parse (flags are populated) and
	// before ctx.Run (commands read tz.Location()).
	if err := cli.ResolveTimezone(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
