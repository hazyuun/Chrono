package cmd

import (
	"chrono/pkg/session"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Session related operations",
	Long:  ``,
}

var sessionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a session",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		s := session.NewSession(repositoryPath)
		s.Start()
	},
}

var sessionStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the session",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal().Msg("Not implemented yet")
	},
}
