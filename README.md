# genact

Version v0.0.4

A golang module and command line programs to interact with Google's
Gemini large context window LLMs.

The two cli programmes provided are:

* `genact`\
  interact with the gemini api, optionally providing history from a
  saved ai studio or api history file, saving history to a structured
  json file together with the prompt and output as noted below.
  The last history file for a chat is used by default.

* `thinner`\
  an inventively-named program for thinning a history file to omit
  user/model conversation pairs to reduce token size and processing
  cost.

## genact

Interact with a Google Gemini large context window LLM API using the
cli.

Simply put, start or continue a chat on the subject of `limericks` as
follows:

```bash
genact -c limericks prompt.txt
```

A timestamped history file will be made by default in
`conversations/limericks` which will be re-used for the next
conversation. Timestamped prompt and output `.md` files will be
similarly saved. Output can be conveniently viewed with markdown readers
such as CharmBracelet's [glow](https://github.com/charmbracelet/glow).

The `genact --help` output is as follows:

```
Usage:
  genact version v0.0.4

Have a conversation with gemini AI with the provided prompt file using
the settings file (by default settings.yaml) and, optionally, either a
history file saved from previous AI discussions or downloaded from
Google AI studio. By default the last-generated history file will be
used, if it exists.

Within "Directory" (by default the current working directory) a
"conversations" directory will be made if it does not exist. For each
chat a further directory will be made, the "chat" directory. Within each
relevant "chat" directory a timestamped prompt and AI response (output)
will be saved, together with the previous history, if any, with the
prompt and response appended. An output.md file is also written from the
response to the current working directory for easy reference.

As a result, a call to this program for the first time along the
following lines:

	./genai -c hi prompt.txt

Will create something like the following output:

	./
	├── conversations
	│   └─── hi
	│       ├── 20250630T204842-history.json
	│       ├── 20250630T204842-output.md
	│       └── 20250630T204842-prompt.txt
	└── output.md

The 20250630T204842-history.json file can be used for the next call to
the api to "continue" the conversation, which is what will happen by
default if no apiHistory or studioHistory is specified.

./genact [-a apiHistory] [-s studioHistory] -c "chat name" \
         [-d directory] [-y yaml]  Prompt

Application Options:
  -a, --apiHistory=    path to api history json file
  -s, --studioHistory= path to studio history json file
  -c, --chatName=      name of this conversation
  -d, --directory=     directory (default: current working directory)
  -y, --yamlFile=      settings yaml file (default: settings.yaml)

Help Options:
  -h, --help           Show this help message

Arguments:
  Prompt:              prompt text file
```

## Licence

This project is licensed under the [MIT Licence](LICENCE).

