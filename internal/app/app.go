package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/upamune/airule/internal/cli"
	"github.com/upamune/airule/internal/tui"
)

// App represents the main application
type App struct {
	cliArgs cli.CLI
}

// NewApp creates a new App instance
func NewApp(cliArgs cli.CLI) *App {
	return &App{
		cliArgs: cliArgs,
	}
}

// Run executes the application
func (a *App) Run() error {
	// Initialize the TUI program
	p := tea.NewProgram(
		tui.InitialModel(a.cliArgs),
		tea.WithAltScreen(),
	)

	// Run the program
	model, err := p.Run()
	if err != nil {
		return err
	}

	// Check for type conversion errors
	finalModel, ok := model.(tui.Model)
	if !ok {
		return fmt.Errorf("could not convert final model to tui.Model")
	}

	// Propagate errors from the model
	if modelErr := finalModel.Error(); modelErr != nil {
		return modelErr
	}

	return nil
}
