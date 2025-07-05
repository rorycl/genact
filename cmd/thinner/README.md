# README

This program, `thinner`, provides an interactive way of thinning a
genesis API history file. This needs the `glow` binary installed on
one's path to read markdown files.

See https://github.com/charmbracelet/glow

The program reads each prompt/response pair (although sometimes there
are more than one response if the AI response includes thinking), which
are dumped to glow in pager mode and then after quitting glow the
program will ask if that history segment should be saved to the thinned
history.

The options for saving are:

y : yes, save the conversation
n : no, don't save the conversation
