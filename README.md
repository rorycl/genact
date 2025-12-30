# genact

Version v0.0.5 (for Gemini 3).

## genact

Interact with a Google Gemini 3 large context window LLM API using the
cli, utilising history and thought signatures. Attachment parsing is
supplorted by the `parsefile` option as a one-shot analysis or as a set
of attachments (at the risk of high token usage counts) to `converse`.

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

## Licence

This project is licensed under the [MIT Licence](LICENCE).

