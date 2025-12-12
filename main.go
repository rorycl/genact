package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"path/filepath"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "genact",
		Version: "v0.1.0-gemini3",
		Usage:   "Interact with Google Gemini 3.0",
		Commands: []*cli.Command{
			// ---------------------------------------------------------
			// CONVERSE
			// ---------------------------------------------------------
			{
				Name:    "converse",
				Aliases: []string{"c"},
				Usage:   "Converse with the model (stateful)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "conversation", Aliases: []string{"c"}, Usage: "Conversation name", Required: true},
					&cli.BoolFlag{Name: "new", Aliases: []string{"n"}, Usage: "Force new conversation"},
					&cli.StringFlag{Name: "history", Aliases: []string{"j", "jsonHistory"}, Usage: "Explicit path to a history file to fork from"},
					&cli.BoolFlag{Name: "old-history", Aliases: []string{"o"}, Usage: "Import from legacy 0.0.4 format"},
					&cli.StringSliceFlag{Name: "attach", Aliases: []string{"a"}, Usage: "Attach files (e.g. -a doc.pdf)"},
					&cli.StringFlag{Name: "thinking", Aliases: []string{"t"}, Value: "high", Usage: "Thinking level (low|high)"},
					&cli.StringFlag{Name: "settings", Aliases: []string{"y"}, Value: "settings.yaml", Usage: "Path to settings file"},
				},
				ArgsUsage: "[prompt_file]",
				Action: func(ctx context.Context, c *cli.Command) error {
					promptPath := c.Args().First()
					if promptPath == "" {
						return fmt.Errorf("prompt file argument is required")
					}

					// 1. Load Settings
					settings, err := LoadYaml(c.String("settings"))
					if err != nil {
						return err
					}
					// Override settings with flags
					settings.ThinkingLevel = c.String("thinking")

					// 2. Setup Paths
					chatName := c.String("conversation")
					paths, err := NewFilePaths("", chatName, false)
					if err != nil {
						return err
					}

					// 3. Resolve History
					var history *HistoryFile
					var forkReason string
					var parentID string

					explicitHistory := c.String("history")
					isLegacy := c.Bool("old-history")
					forceNew := c.Bool("new")

					if forceNew {
						// Start fresh
						forkReason = "new"
					} else if explicitHistory != "" {
						// Fork from specific file
						if isLegacy {
							history, err = MigrateLegacy(explicitHistory)
							if err != nil {
								return err
							}
							forkReason = "legacy_fork"
						} else {
							history, err = LoadHistory(explicitHistory)
							if err != nil {
								return err
							}
							forkReason = "branch"
						}
						// If forking, we define the PARENT as the ID of the loaded file
						parentID = history.ID
					} else {
						// Auto-continue latest
						latest, err := FindLatestHistory(paths.ChatDir)
						if err != nil {
							return err
						}
						if latest != "" {
							history, err = LoadHistory(latest)
							if err != nil {
								return fmt.Errorf("failed to load previous history: %w", err)
							}
							parentID = history.ID
							forkReason = "reply"
						} else {
							forkReason = "new"
						}
					}

					// 4. Create New ID for this session
					sessionID := paths.Timestamp

					// If history exists, we technically prepare a 'new' history object based on it
					// but we only populate the Previous Turns for the API call.
					if history == nil {
						history = &HistoryFile{
							ID:            sessionID,
							ParentID:      "",
							ForkReason:    "new",
							OriginalModel: settings.ModelName,
							Turns:         []Turn{},
						}
					} else {
						// Prepare for next step
						// We don't modify 'history' in place yet, we pass it to API
					}

					// 5. Prepare User Turn
					promptBytes, err := os.ReadFile(promptPath)
					if err != nil {
						return err
					}

					userTurn := Turn{
						Role:      "user",
						TextParts: []string{string(promptBytes)},
						Timestamp: time.Now(),
					}

					// Load attachments
					attachPaths := c.StringSlice("attach")
					for _, ap := range attachPaths {
						att, err := LoadAttachment(ap)
						if err != nil {
							return err
						}
						userTurn.Attachments = append(userTurn.Attachments, att)
					}

					// 6. API Call
					fmt.Printf("Generating response using %s...\n", settings.ModelName)
					modelTurn, tokenCount, err := GenerateResponse(ctx, settings, history, userTurn)
					if err != nil {
						return err
					}

					// 7. Update History & Save
					// We create a NEW history file that contains:
					// [Ancestors] + [UserTurn] + [ModelTurn]
					newHistory := history.DeepCopy(sessionID, forkReason)
					newHistory.ParentID = parentID
					newHistory.OriginalModel = settings.ModelName
					newHistory.TotalTokens = tokenCount
					newHistory.Turns = append(newHistory.Turns, userTurn, *modelTurn)

					historyBytes, _ := newHistory.Serialize()
					if err := os.WriteFile(paths.HistoryFile, historyBytes, 0644); err != nil {
						return err
					}

					// 8. Write Outputs
					fullOutput := strings.Join(modelTurn.TextParts, "\n\n")
					if modelTurn.Thought != "" {
						fullOutput = fmt.Sprintf("<details><summary>Thinking</summary>\n\n%s\n\n</details>\n\n%s", modelTurn.Thought, fullOutput)
					}

					os.WriteFile(paths.ResponseFile, []byte(fullOutput), 0644)
					os.WriteFile(paths.PromptFile, promptBytes, 0644) // Save snapshot of prompt

					fmt.Printf("Done. Saved to %s\n", paths.HistoryFile)
					return nil
				},
			},

			// ---------------------------------------------------------
			// REGEN / STRIP
			// ---------------------------------------------------------
			{
				Name:    "regen",
				Aliases: []string{"strip", "r"},
				Usage:   "Regenerate a history file with no thought signatures",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "apiHistory", Aliases: []string{"a"}, Required: true, Usage: "History file to strip"},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					inputPath := c.String("apiHistory")
					history, err := LoadHistory(inputPath)
					if err != nil {
						return err
					}

					// Perform stripping
					StripSignatures(history)

					// Update Metadata
					ts := time.Now().Format(timeFormat)
					history.ParentID = history.ID
					history.ID = ts
					history.ForkReason = "strip_signatures"

					// Save to same dir with new timestamp
					dir := filepath.Dir(inputPath)
					newPath := filepath.Join(dir, fmt.Sprintf("%s_history.json", ts))

					bytes, _ := history.Serialize()
					if err := os.WriteFile(newPath, bytes, 0644); err != nil {
						return err
					}

					fmt.Printf("Stripped signatures. New file: %s\n", newPath)
					return nil
				},
			},

			// ---------------------------------------------------------
			// PARSEFILE
			// ---------------------------------------------------------
			{
				Name:    "parsefile",
				Aliases: []string{"p"},
				Usage:   "One-shot analysis of files (saved to files/ directory)",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "file", Aliases: []string{"f"}, Required: true, Usage: "Files to attach"},
					&cli.StringFlag{Name: "settings", Aliases: []string{"y"}, Value: "settings.yaml", Usage: "Path to settings file"},
				},
				ArgsUsage: "[prompt_file]",
				Action: func(ctx context.Context, c *cli.Command) error {
					promptPath := c.Args().First()
					if promptPath == "" {
						return fmt.Errorf("prompt file argument is required")
					}

					settings, err := LoadYaml(c.String("settings"))
					if err != nil {
						return err
					}

					// Setup "files" directory structure
					paths, err := NewFilePaths("", "parsed", true) // "parsed" is a generic bucket name in files/
					if err != nil {
						return err
					}

					// Prepare User Turn (No History)
					promptBytes, err := os.ReadFile(promptPath)
					if err != nil {
						return err
					}

					userTurn := Turn{
						Role:      "user",
						TextParts: []string{string(promptBytes)},
						Timestamp: time.Now(),
					}

					for _, f := range c.StringSlice("file") {
						att, err := LoadAttachment(f)
						if err != nil {
							return err
						}
						userTurn.Attachments = append(userTurn.Attachments, att)
					}

					fmt.Println("Parsing file(s)...")
					// Pass nil history
					modelTurn, _, err := GenerateResponse(ctx, settings, nil, userTurn)
					if err != nil {
						return err
					}

					// Write Output
					fullOutput := strings.Join(modelTurn.TextParts, "\n\n")
					if err := os.WriteFile(paths.ResponseFile, []byte(fullOutput), 0644); err != nil {
						return err
					}

					// We DO NOT save history.json for parsefile, as requested,
					// but we do save the response markdown.
					fmt.Printf("Output saved to %s\n", paths.ResponseFile)
					return nil
				},
			},

			// ---------------------------------------------------------
			// LINEAGE
			// ---------------------------------------------------------
			{
				Name:    "lineage",
				Aliases: []string{"l"},
				Usage:   "View the lineage of a conversation",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "conversation", Aliases: []string{"c"}, Required: true},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					chatName := c.String("conversation")
					paths, err := NewFilePaths("", chatName, false)
					if err != nil {
						return err
					}

					files, err := ScanForLineage(paths.ChatDir)
					if err != nil {
						return err
					}

					if len(files) == 0 {
						fmt.Println("No history found.")
						return nil
					}

					// Simple output
					fmt.Printf("Lineage for %s:\n", chatName)
					fmt.Println("ID                  | Parent              | Reason             | Model          | Tokens")
					fmt.Println("---------------------------------------------------------------------------------------")
					for _, f := range files {
						parent := f.ParentID
						if len(parent) > 15 {
							parent = "..." + parent[len(parent)-12:]
						}
						if parent == "" {
							parent = "ROOT"
						}

						fmt.Printf("%-19s | %-19s | %-18s | %-14s | %d\n",
							f.ID, parent, f.ForkReason, f.OriginalModel, f.TotalTokens)
					}

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
