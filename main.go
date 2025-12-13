package main

import (
	"context"
	"fmt"
	"os"
)

func main() {

	app := NewApp()
	cmd := BuildCLI(app)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
