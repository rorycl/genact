# genact

Version v0.0.3

A golang module and command line programs to interact with Google's
Gemini large context window LLMs.

The two cli programmes provided are:

* `genact`\
  interact with the gemini api, optionally providing history from a
  saved ai studio or api history file, saving history to a structured
  json file together with the prompt and output as noted below.

* `thinner`\
  an inventively-named program for thinning a history file to omit
  user/model conversation pairs to reduce token size and processing
  cost.

## genact

Interact with a Google Gemini large context window LLM API over the cli.

```
Usage:
  genact version v0.0.3

Have a conversation with gemini AI with the provided prompt file using
the settings file (by default at settings.yaml) and, optionally, either
a history file saved from previous AI discussions or downloaded from
Google AI studio.

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
the api to "continue" the conversation.

./genact [-a apiHistory] [-s studioHistory] -c "chat name" \
         [-d directory] [-y yaml] Prompt

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

## thinner

`thinner` thins a history file saved from previous interactions using
`genact` to remove unneeded items to reduce token size and therefore
cost.


```
Usage:
  thinner version v0.0.3

Interactively "thin" a gemini history file saved with genact by choosing
which conversations from the history to output to a new history json file.

Note that the conversations are replayed in reverse order, but
recompiled in the original order.

This uses bubbletea's "glow" markdown pager programme, which needs to be
on your PATH.

Using the -r/--review flag only reviews the (0-indexed) conversations
numbered, the other indexed items are kept. Negative indexing can be
used to refer to items from the end of the list of conversations, so -1
means the last item.

Usint the -k/--keep flag presets the items to keep. This may be used in
combination with the -r/--review items which may be different or
overlapping sets, where at most the -k + -r conversations will be kept
after interactive review.

./thinner [-o outputFile] [-r 1, -r 3...] [-k 0, -k 2...] InputFile

Application Options:
  -o, --outputFile= file path to save output
  -r, --review=     list of specific conversation pairs to review
  -k, --keep=       list of specific conversation pairs to keep

Help Options:
  -h, --help        Show this help message

Arguments:
  InputFile:        input history json file

```

## Licence

This project is licensed under the [MIT Licence](LICENCE).
