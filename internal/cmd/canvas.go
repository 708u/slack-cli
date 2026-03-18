package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// CanvasCmd groups canvas subcommands.
type CanvasCmd struct {
	Read CanvasReadCmd `cmd:"" help:"Read canvas sections"`
	List CanvasListCmd `cmd:"" help:"List canvases in channel"`
}

// CanvasReadCmd reads sections of a canvas.
type CanvasReadCmd struct {
	ID      string `name:"id" short:"i" required:"" help:"Canvas ID"`
	Format  string `name:"format" optional:"" default:"table" enum:"table,simple,json" help:"Output format: table, simple, json"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the canvas read command.
func (c *CanvasReadCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	sections, err := client.ReadCanvas(c.ID)
	if err != nil {
		return err
	}

	if len(sections) == 0 {
		fmt.Println("No sections found in canvas")
		return nil
	}

	f := format.ParseFormat(c.Format)
	switch f {
	case format.JSON:
		format.PrintJSON(sections)
	case format.Simple:
		for _, s := range sections {
			id := s.ID
			if id == "" {
				id = "(no id)"
			}
			text := extractCanvasText(s.Elements)
			if text == "" {
				text = "(no content)"
			}
			fmt.Printf("%s\t%s\n", id, text)
		}
	default:
		cyan := color.New(color.FgCyan)
		for _, s := range sections {
			id := s.ID
			if id == "" {
				id = "(no id)"
			}
			text := extractCanvasText(s.Elements)
			if text == "" {
				text = "(no content)"
			}
			fmt.Printf("%s  Content: %s\n", cyan.Sprintf("ID: %s", id), text)
		}
	}
	return nil
}

// CanvasListCmd lists canvases in a channel.
type CanvasListCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Format  string `name:"format" optional:"" default:"table" enum:"table,simple,json" help:"Output format: table, simple, json"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the canvas list command.
func (c *CanvasListCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	canvases, err := client.ListCanvases(c.Channel)
	if err != nil {
		return err
	}

	if len(canvases) == 0 {
		fmt.Println("No canvases found in channel")
		return nil
	}

	f := format.ParseFormat(c.Format)
	switch f {
	case format.JSON:
		format.PrintJSON(canvases)
	case format.Simple:
		for _, cv := range canvases {
			id := cv.ID
			if id == "" {
				id = "(no id)"
			}
			name := cv.Name
			if name == "" {
				name = "(no name)"
			}
			fmt.Printf("%s\t%s\n", id, name)
		}
	default:
		cyan := color.New(color.FgCyan)
		for _, cv := range canvases {
			id := cv.ID
			if id == "" {
				id = "(no id)"
			}
			name := cv.Name
			if name == "" {
				name = "(no name)"
			}
			fmt.Printf("%s  Name: %s\n", cyan.Sprintf("ID: %s", id), name)
		}
	}
	return nil
}

// extractCanvasText recursively extracts text from canvas section elements.
func extractCanvasText(elements []slack.CanvasSectionElement) string {
	var result string
	for _, el := range elements {
		if el.Text != "" {
			result += el.Text
		}
		if len(el.Elements) > 0 {
			result += extractCanvasText(el.Elements)
		}
	}
	return result
}
