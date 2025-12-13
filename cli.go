package main

import (
	"context"
	"genact/app"

	"github.com/urfave/cli/v3"
)

const (
	ShortUsage      = "A cli program for interacting with Gemini 3 series LLMs"
	LongDescription = `The genact program uses a local settings and
   prompt file to 'converse' with the LLM. Additional facilities are
   provided for one-shot file parsing, history regeneration and history
   lineage reporting.`
)

// Applicator is an interface to the central coordinator for the project
// (concretely provided by App in app.go) to allow for testing.
type Applicator interface {
	Converse(ctx context.Context, cfg app.ConverseOptions) error
	Regenerate(inputPath string) error
	ParseFiles(ctx context.Context, settingsPath, promptPath string, attachments []string) error
	Lineage(conversation string) error
}

// BuildCLI creates a cli app to run the capabilities provided by
// an Applicator dependency.
//
// This is work in progress.
func BuildCLI(a Applicator) *cli.Command {

	// Converse
	converseCmd := &cli.Command{
		Name:  "converse",
		Usage: "Converse with gemini 3.0 LLMs",
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
				Usage:   "Make a new conversation",
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
				Name:    "settings",
				Aliases: []string{"y"},
				Value:   "settings.yaml",
				Usage:   "Path to settings file (required)"},
		},
		ArgsUsage: "`PROMPT_FILE`",

		Action: func(ctx context.Context, c *cli.Command) error {
			return a.Converse(
				ctx,
				app.ConverseOptions{
					ConversationName: c.String("conversation"),
					PromptPath:       c.Args().First(),
					SettingsPath:     c.String("settings"),
					HistoryPath:      c.String("history"),
					ThinkingLevel:    c.String("thinking"),
					IsNew:            c.Bool("new"),
					IsLegacyHistory:  c.Bool("old-history"),
					Attachments:      c.StringSlice("attach"),
				},
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
			return a.Regenerate(
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
			return a.ParseFiles(
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
			return a.Lineage(
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
