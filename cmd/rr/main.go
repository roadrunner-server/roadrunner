package main

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/roadrunner-server/roadrunner/v2/internal/cli"
)

// exitFn is a function for application exiting.
var exitFn = os.Exit //nolint:gochecknoglobals

// main CLI application entrypoint.
func main() { exitFn(run()) }

// run this CLI application.
func run() int {
	cmd := cli.NewCommand(filepath.Base(os.Args[0]))

	if err := cmd.Execute(); err != nil {
		_, _ = color.New(color.FgHiRed, color.Bold).Fprintln(os.Stderr, err.Error())

		return 1
	}

	return 0
}
