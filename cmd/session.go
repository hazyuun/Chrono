package cmd

import (
	"chrono/pkg/chrono"
	"chrono/pkg/chrono/session"
	"errors"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Session related operations",
	Long:  ``,
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creates a new session",
	Long:  ``,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please specify a session name")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		session.CreateSession(args[0])
		log.Info().Str("session", args[0]).Msg("Session created successfully")
	},
}

var sessionDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a session",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		session.DeleteSession(args[0])
		log.Info().Str("session", args[0]).Msg("Session deleted successfully")
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists existing sessions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		sessions := session.GetSessions()

		tbl := table.New("N°", "Session name", "Chrono branch", "Source Branch")

		tbl.WithHeaderFormatter(color.New(color.FgBlue, color.Underline, color.Bold).SprintfFunc())
		tbl.WithFirstColumnFormatter(color.New(color.FgYellow, color.Bold).SprintfFunc())
		tbl.WithPadding(8)

		i := 1
		for _, session := range sessions {
			tbl.AddRow(i, session.Name, session.Branch, session.Source)
			i++
		}

		tbl.Print()
	},
}

var sessionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a session",
	Long:  ``,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please specify a session name")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		s := session.OpenSession(args[0])
		log.Info().Str("session", args[0]).Msg("Session opened")
		s.Start()
	},
}

var sessionMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "To squash merge all session commits to the original branch",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please specify a session name")
		} else if len(args) < 2 {
			return errors.New("Please provide a commit message")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		s := session.OpenSession(args[0])
		log.Info().Str("session", args[0]).Msg("Session opened")
		s.SquashMerge(args[1])
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

var sessionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows git commits specific to a session",
	Long:  ``,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please specify a session name")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		chrono.Init(repositoryPath)
		commits := session.GetSessionCommits(args[0])

		tbl := table.New("Hash", "Author", "Message")

		tbl.WithHeaderFormatter(color.New(color.FgBlue, color.Underline, color.Bold).SprintfFunc())
		tbl.WithFirstColumnFormatter(color.New(color.FgYellow, color.Bold).SprintfFunc())
		tbl.WithPadding(8)

		for _, c := range commits {
			tbl.AddRow(c.Hash[:8], c.Author, c.Message)
		}

		tbl.Print()
	},
}
