package main

import (
	"fmt"
	"os"

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
