package cmd

import (
	"github.com/richer421/q-metahub/infra/mysql"
	"github.com/richer421/q-metahub/pkg/logger"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mysql.Init(); err != nil {
			logger.Fatalf("mysql init failed: %v", err)
		}
		defer mysql.Close()

		logger.Infof("Migration completed!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
