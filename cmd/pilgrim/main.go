package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/williams-jack/pilgrim/internal/config"
	"github.com/williams-jack/pilgrim/internal/migrations"
	"github.com/williams-jack/pilgrim/internal/postgres"
)

const version = "0.0.1"

var (
	configPath    string
	createDirs    bool
	pilgrimConfig *config.PilgrimConfig
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "pilgrim",
	Short: "A migration tool for databases",
	Long:  "Pilgrim is a database migration tool that makes managing schema changes slightly easier.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configFile, err := os.Open(configPath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		config, err := config.ReadFromReader(configFile)
		if err != nil {
			return err
		}
		pilgrimConfig = config
		return nil
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [migration_name]",
	Short: "Rollback migrations up to and including the specified migration name",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var migrationName string
		if len(args) == 1 {
			migrationName = args[0]
		}
		// TODO: Eventually configure variable timeouts for connection, queries,
		// and internal operations.
		// For now, we will use default timeouts (by providing no args).
		rollbackArgs := &postgres.RollbackArgs{
			MigrationName:    migrationName,
			ConnectionString: pilgrimConfig.ConnectionString(),
			DownDir:          pilgrimConfig.DownDir,
		}
		rolledBackMigrations, err := postgres.RollbackToMigration(rollbackArgs)
		if err != nil {
			return err
		}
		for _, m := range rolledBackMigrations {
			fmt.Printf("Rolled back migration: %s\n", m)
		}
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply pending migrations to the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch pilgrimConfig.DbType {
		case "postgres":
			applyArgs := &postgres.ApplyMigrationsArgs{
				ConnectionString: pilgrimConfig.ConnectionString(),
				UpDir:            pilgrimConfig.UpDir,
			}
			appliedMigrations, err := postgres.ApplyMigrations(applyArgs)
			for _, m := range appliedMigrations {
				fmt.Printf("Applied migration: %s\n", m)
			}
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported dbType: %s", pilgrimConfig.DbType)
		}
		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the migration tracking table in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch pilgrimConfig.DbType {
		case "postgres":
			err = postgres.PgInitMigrationTable(pilgrimConfig.ConnectionString())
		default:
			return fmt.Errorf("unsupported dbType: %s", pilgrimConfig.DbType)
		}
		if err == nil && verbose {
			fmt.Println("Migration tracking table initialized successfully.")
		}
		return err
	},
}

var createCmd = &cobra.Command{
	Use:   "create name",
	Short: "Create a new migration files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName := args[0]
		err := migrations.CreateMigration(fileName,
			pilgrimConfig.UpDir,
			pilgrimConfig.DownDir,
			pilgrimConfig.DbType,
			createDirs)
		if err == nil && verbose {
			fmt.Printf("Created migration files for: %s\n", fileName)
		}
		return err
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pilgrim",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version)
	},
}

func init() {
	createCmd.Flags().BoolVarP(&createDirs, "create-dirs", "d", false, "Create directories for the migration files")
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "pilgrim.config.json", "Path to the configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
