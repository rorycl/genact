package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	ShortUsage      = "A cli program for interacting with Genesis LLMs"
	LongDescription = `The genact program uses a local settings and
   prompt file to 'converse' with the LLM. Additional facilities are
   provided for one-shot file parsing, history regeneration and history
   lineage reporting.`
)

// Applicator is an interface to the central coordinator for the project
// (concretely provided by App in app.go) to allow for testing.
type Applicator interface {
	Converse(ctx context.Context, conversationName, promptPath, settingsPath, historyPath, thinkingLevel string, isNew, isHistoric bool, attachments []string) error
	Regenerate(inputPath string) error
	ParseFiles(ctx context.Context, settingsPath, promptPath string, attachments []string) error
	Lineage(conversation string) error
}

// BuildCLI creates a cli app to run the capabilities provided by
// an Applicator dependency.
//
// This is work in progress.
func BuildCLI(app Applicator) *cli.Command {

	// isFile determines if a file exists.
	isFile := func(path string) bool {
		o, err := os.Stat(path)
		if err != nil {
			return false
		}
		if o.IsDir() == true {
			return false
		}
		return true
	}

	// Return true if a path is empty ("") or if it is not empty, check
	// if the file exists.
	isNotEmptyAndIsFile := func(path string) bool {
		if path == "" {
			return false
		}
		return isFile(path)
	}

	// Converse
	converseCmd := &cli.Command{
		Name:  "converse",
		Usage: "Converse with gemini",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conversation",
				Aliases:  []string{"c"},
				Usage:    "Conversation name",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Ensure the conversation is new",
			},
			&cli.StringFlag{
				Name:    "history",
				Aliases: []string{"j", "jsonHistory"},
				Usage:   "Explicit path to a history file to fork from",
			},
			&cli.BoolFlag{
				Name:    "old-history",
				Aliases: []string{"o"},
				Usage:   "Import from legacy 0.0.4 format",
			},
			&cli.StringSliceFlag{
				Name:    "attach",
				Aliases: []string{"a"},
				Usage:   "Attach files (e.g. -a doc.pdf)",
			},
			&cli.StringFlag{Name: "thinking",
				Aliases: []string{"t"},
				Value:   "high",
				Usage:   "Override settings thinking level (low|high)",
			},
			&cli.StringFlag{
				Name:     "settings",
				Aliases:  []string{"y"},
				Value:    "settings.yaml",
				Required: true,
				Usage:    "Path to settings file (required)"},
		},
		ArgsUsage: "`PROMPT_FILE`",

		// Before runs validations before "Action" is run.
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.NArg() != 1 {
				return ctx, fmt.Errorf("missing required argument: `PROMPT_FILE`")
			}
			promptFile := c.Args().First()
			if isFile(promptFile) == false {
				return ctx, fmt.Errorf("prompt file %q not found", promptFile)
			}

			conversation := c.String("conversation")
			settings := c.String("settings")
			// newFlag := c.Bool("new")
			historyFile := c.String("history")
			// isHistoricHistory := c.Bool("old-history")
			attachments := c.StringSlice("attach")

			if conversation == "" {
				return ctx, errors.New("conversation name `-c` cannot be empty")
			}

			// File existence checks.
			if isFile(settings) == false {
				return ctx, fmt.Errorf("settings file %q not found", promptFile)
			}
			if isNotEmptyAndIsFile(historyFile) {
				return ctx, fmt.Errorf("history file %q not found", promptFile)
			}
			for _, a := range attachments {
				if isFile(a) == false {
					return ctx, fmt.Errorf("attachment %q not found", a)
				}
			}
			return ctx, nil
		},

		Action: func(ctx context.Context, c *cli.Command) error {

			return app.Converse(
				ctx,
				c.String("conversation"),
				c.Args().First(), // prompt file
				c.String("settings"),
				c.String("history"),
				c.String("thinking"),
				c.Bool("new"),
				c.Bool("old-history"),
				c.StringSlice("attach"),
			)
		},
	}

	// regen
	regenerationCmd := &cli.Command{
		Name:    "regen",
		Aliases: []string{"strip"},
		Usage:   "Regenerate a history file with no thought signatures",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "apiHistory", Aliases: []string{"a"}, Required: true, Usage: "History file to strip"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.Regenerate(
				c.String("apiHistory"),
			)
		},
	}

	// parsefile
	parsefileCmd := &cli.Command{
		Name:  "parsefile",
		Usage: "One-shot analysis of files (saved to files/ directory)",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "file", Aliases: []string{"f"}, Required: true, Usage: "Files to attach"},
			&cli.StringFlag{Name: "settings", Aliases: []string{"y"}, Value: "settings.yaml", Usage: "Path to settings file"},
		},
		ArgsUsage: "[prompt_file]",
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.ParseFiles(
				ctx,
				c.String("settings"),
				c.Args().First(), // prompt file
				c.StringSlice("file"),
			)
		},
	}

	// lineage
	lineageCmd := &cli.Command{
		Name:  "lineage",
		Usage: "View the lineage of a conversation",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "conversation", Aliases: []string{"c"}, Required: true},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return app.Lineage(
				c.String("conversation"),
			)
		},
	}

	rootCmd := &cli.Command{
		Name:        "genact",
		Version:     "v0.1.0-gemini3",
		Usage:       ShortUsage,
		Description: LongDescription,
		Commands:    []*cli.Command{converseCmd, regenerationCmd, parsefileCmd, lineageCmd},
	}

	// custom help template.
	rootCmd.CustomRootCommandHelpTemplate = rootHelpTemplate

	return rootCmd

}

var rootHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [command] [options]

DESCRIPTION:
   {{.Description}}

COMMANDS:
{{range .Commands}}   {{.Name}}{{ "\t"}}{{.Usage}}
{{end}}
Run '{{.Name}} [command] --help' for more information on a command.`
