package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configPath = ".awsSts"
	configName = "config"
	configExt  = "json"
)

func Run() {
	pflag.String("username", "", "Username for your Single Sign On")
	pflag.String("password", "", "Password for your Single Sign On")
	pflag.String("url", "", "URL for your Single Sign On")

	pflag.StringP("profile", "p", "default", "AWS profile to store temp credentials against")
	pflag.StringP("role", "r", "", "Role to auto select, if one one role available it is auto selected")

	pflag.Bool("token", false, "If set the token is displayed (useful for tools that don't use aws credentials file)")

	pflag.Bool("verbose", false, "Display details of internal process")
	ver := pflag.Bool("version", false, "Display version info")
	pflag.Bool("dump-work", false, "For HTTP requests save local copies of responses")

	pflag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stdout, "Using your Single Sign On credentials a temporary token will be created and stored in your .aws credentials file.")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if *ver {
		fmt.Printf("%v", _VERSION)
		return
	}

	initViper()

	viper.BindPFlags(pflag.CommandLine)

	execLogon()
}

// initConfig reads in config file and ENV variables if set.
func initViper() {
	//viper.Set("verbose", false)   //, ""
	//viper.Set("dump-work", false) //, "")

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
