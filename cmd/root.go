package cmd

import (
	"fmt"

	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"chrono/pkg/config"
	"chrono/pkg/signal"
)

var logFile string
var repositoryPath string

func setupLogger() {
	if logFile == "" {
		log.Logger = log.Output(
			&zerolog.ConsoleWriter{Out: os.Stderr},
		)
		return
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Couldn't open log file : %v", err.Error())
		os.Exit(1)
	}

	log.Logger = log.Output(
		zerolog.MultiLevelWriter(
			&zerolog.ConsoleWriter{Out: os.Stderr},
			f,
		),
	)
}

var rootCmd = &cobra.Command{
	Use:   "chrono",
	Short: "Chrono is a git time machine",
	Long: `	Chrono is a tool that automatically commits in a temporary branch 
	in your git repository every time an event occurs (events are customizable),
	So that you can always rollback to a specific point in time if anything goes wrong. 
	You can squash merge all the temporary commits into one once you are done.`,
}

func Run() error {
	return rootCmd.Execute()
}

func init() {
	signal.Init()
	setupLogger()

	cobra.OnInitialize(config.Load)
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", "Log file path")
	rootCmd.PersistentFlags().StringVarP(&repositoryPath, "repository", "r", ".", "Git repository path")

	rootCmd.AddCommand(sessionCmd)

	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCmd.AddCommand(sessionDeleteCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionStartCmd)
	sessionCmd.AddCommand(sessionStopCmd)
	sessionCmd.AddCommand(sessionMergeCmd)
	sessionCmd.AddCommand(sessionShowCmd)
}
