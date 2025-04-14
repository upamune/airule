package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/upamune/airule/internal/app"
	"github.com/upamune/airule/internal/cli"
)

const appName = "airule"

func main() {
	var cliArgs cli.CLI
	parser, err := kong.New(&cliArgs,
		kong.Name(appName),
		kong.Description("Interactively copy rule files."),
		kong.Vars{"version": fmt.Sprintf("%s %s", appName, cli.GetVersion())},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing CLI: %v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	if _, err := parser.Parse(os.Args[1:]); err != nil {
		parser.FatalIfErrorf(err)
	}

	// Validate CLI arguments
	if err := cliArgs.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run the application
	app := app.NewApp(cliArgs)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
