package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mattes/migrate/migrate"
	_ "github.com/mattes/migrate/driver/postgres"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the database",
	Run:   migrateDB,
}

func init() {
	RootCmd.AddCommand(MigrateCmd)
}

func migrateDB(_ *cobra.Command, _ []string) {
	migrate.UpSync("postgres://postgres:postgres@mg4:5432", "./db/migrations")
}
