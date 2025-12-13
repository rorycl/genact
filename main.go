package main

import (
	"context"
	"fmt"
	"os"

	"genact/app"
)

func main() {

	a := app.NewApp()
	cmd := BuildCLI(a)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
