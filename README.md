# genact

Version v0.0.5 (for Gemini 3.1).

## genact

Interact with a Google Gemini 3 large context window LLM API using the
cli, utilising history and thought signatures. Attachment parsing is
supplorted by the `parsefile` option as a one-shot analysis or as a set
of attachments (at the risk of high token usage counts) to `converse`.

The lineage subcommand is useful for tracking history forks.

```
NAME:
   genact - A cli program for interacting with Gemini 3 series LLMs

USAGE:
   genact [command] [options]

DESCRIPTION:
   The genact program uses a local settings and prompt file to
   'converse' with the LLM. Additional facilities are provided for
   one-shot file parsing, history regeneration and history lineage
   reporting.

COMMANDS:
   converse   Converse with gemini 3.0 LLMs
   regen      Regenerate a history file with no thought signatures
   parsefile  One-shot analysis of files (saved to files/ directory)
   lineage    View the lineage of a conversation
   help       Shows a list of commands or help for one command

Run 'genact [command] --help' for more information on a command.
```

Settings are something like the following `settings.example.yaml`:

```yaml
---
modelName  : "gemini-3.1-pro-preview"
apiKey     : "YOuR_API_KEY"
logging    : true

# A set of golang regexps check inputs are properly anonymised.
anonymisationChecks:
  - 'rabbit'
  - '(?i)\bBattle.*Bulge\b'
```

The main subcommand is `converse` which by default checks for personal
data in the prompt or attachments based on the set of
`anonymisationChecks` regexes in the settings file.

```
NAME:
   genact converse - Converse with gemini 3.0 LLMs

USAGE:
   genact converse [options] `PROMPT_FILE`

OPTIONS:
   --conversation string, -c string                           Conversation name
   --new, -n                                                  Make a new conversation
   --history string, -j string, --jsonHistory string          Explicit path to a history file to fork from
   --old-history, -o                                          Import from legacy 0.0.4 format
   --attach string, -a string [ --attach string, -a string ]  Attach files (e.g. -a doc.pdf)
   --skipAnonymisationChecks, -s                              Skip anonymisation checks
   --thinking string, -t string                               Override settings thinking level (low|high) (default: "high")
   --settings string, -y string                               Path to settings file (required) (default: "settings.yaml")
   --help, -h                                                 show help
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).

