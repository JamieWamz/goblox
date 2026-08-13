package cli

import (
	"fmt"
	"os"

	"github.com/JamieWamz/goblox/internal/storage"
	"github.com/spf13/cobra"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "goblox",
	Short: "goblox - CLI task tracker",
	Long: `goblox is a production-grade task management tool for your terminal.
Manage your tasks with priority, due dates, and status tracking.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./goblox.db", "path to database file")
}

func getDB() (*storage.Database, error) {
	return storage.NewDatabase(dbPath)
}
