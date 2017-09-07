package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	configPath = ".awsSts"
	configName = "config"
	configExt  = "json"
)

const (
	urlEnv  = "AWSSTS_URL"
	userEnv = "AWSSTS_USER"
	passEnv = "AWSSTS_PASS"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "awsSts2",
	Short: "Small AWS toolkit",
	Long:  `Prime useage is to allow single sign on session for CLI`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/%s/%s.%s)", configPath, configName, configExt))
	RootCmd.PersistentFlags().Bool("verbose", false, "Display details of internal process")
	RootCmd.PersistentFlags().Bool("dump-work", false, "For HTTP requests save local copies of responses")

	viper.BindPFlags(RootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(configName) // name of config file (without extension)
	viper.SetConfigType(configExt)
	viper.AddConfigPath(path.Join("$HOME", configPath)) // adding home directory as first search path
	viper.AutomaticEnv()                                // read in environment variables that match
	viper.SetEnvPrefix("awssts")

	// If a config file is found, read it in.
	for i := 0; i <= 1; i++ {
		if err := viper.ReadInConfig(); err == nil {
			journal("Using config file:'%s'", viper.ConfigFileUsed())
			break
		}
		createInitialConfig()
	}
}

func createInitialConfig() {
	path := filepath.Join(homeDirectory(), configPath, configName+"."+configExt)

	fileMustExist(path)

	c := struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Profile  string `json:"profile"`
	}{
		URL:      os.Getenv("AWSSTS_URL"),
		Username: os.Getenv("AWSSTS_USER"),
		Profile:  "default",
	}
	// auto upgrade

	body, _ := json.MarshalIndent(c, "", "  ")
	if err := ioutil.WriteFile(path, body, os.ModeAppend); err != nil {
		journal("failed to write cfg:%v\n", err)
	}
}
