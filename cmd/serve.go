package cmd

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the telegram bot",
	Run:   serve,
}

func serve(_ *cobra.Command, _ []string) {
	_ = godotenv.Load()
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
