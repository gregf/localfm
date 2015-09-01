package commands

import (
	"fmt"
	"log"

	"github.com/gregf/localfm/src/database"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Podfetcher Version number
const localFMVersion = "v0.1"

// Env struct
type Env struct {
	db database.Datastore
}

// Execute parses command line args and fires up commands
func Execute() {
	initConfig()

	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	env := &Env{db}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Hugo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("LocalFM: %s\n", localFMVersion)
		},
	}

	var cmdImport = &cobra.Command{
		Use:   "import",
		Short: "Imports all your listened to tracks from lastfm",
		Run:   env.Import,
	}

	var cmdDaemon = &cobra.Command{
		Use:   "daemon",
		Short: "Run as a daemon importing data from lastfm",
		Run:   env.Daemon,
	}

	var cmdStats = &cobra.Command{
		Use:   "stats",
		Short: "Display statistics about your LocalFM data",
		Run:   env.Stats,
	}

	var rootCmd = &cobra.Command{Use: "localfm"}
	rootCmd.AddCommand(
		cmdImport,
		cmdDaemon,
		cmdStats,
		cmdVersion)
	rootCmd.Execute()
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/localfm")
	viper.AddConfigPath("$HOME/.config/localfm")
	viper.AddConfigPath("$XDG_CONFIG_HOME/localfm")
	viper.SetConfigType("yml")

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Fatal error config file %s\n", err)
	}
}
