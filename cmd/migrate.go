package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mattes/migrate/migrate"
	_ "github.com/mattes/migrate/driver/postgres"
	"fmt"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the database",
	Run:   migrateDB,
}

var migrationsPath string

func init() {
	RootCmd.AddCommand(MigrateCmd)
	MigrateCmd.Flags().StringVarP(&migrationsPath, "path", "m", "./db/migrations", "The path for migrations")
}

func migrateDB(_ *cobra.Command, _ []string) {
	allErrors, ok := migrate.UpSync("postgres://postgres:postgres@db:5432/mg4?sslmode=disable", migrationsPath)
	if !ok {
		for _, err := range allErrors {
			fmt.Println(err.Error())
		}
	}
}
