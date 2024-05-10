package cmd

import (
	"embed"
	"fmt"

	"github.com/catalystsquad/app-utils-go/env"
	"github.com/catalystsquad/app-utils-go/errorutils"
	"github.com/catalystsquad/app-utils-go/logging"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS
var uri = env.GetEnvOrDefault("POSTGRES_URI", "postgresql://postgres:postgres@localhost:5432?sslmode=disable")

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Performs taikai database migrations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runMigrations() {
	fmt.Println("migrate called")
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	errorutils.PanicOnErr(nil, "error opening database connection", err)
	sqldb, err := db.DB()
	errorutils.PanicOnErr(nil, "error getting database connection", err)
	// set goose file system to use the embedded migrations
	goose.SetBaseFS(migrations)
	logging.Log.Info("Running migrations")
	err = goose.Up(sqldb, "migrations")
	errorutils.PanicOnErr(nil, "error running migrations", err)
}