package main

import (
	"fmt"
	"os"

	"github.com/708u/slack-cli/cmd"
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
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
