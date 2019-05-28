package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logonCmd represents the logon command
var logonCmd = &cobra.Command{
	Use:   "logon",
	Short: "Logon and provide temporary tokens",
	Long:  `Using your Single Sign On credentials a temporary token will be created and stored in your .aws credentials file.`,
	Run:   execLogon,
}

func init() {
	RootCmd.AddCommand(logonCmd)

	logonCmd.Flags().String("username", "", "Username for your Single Sign On")
	logonCmd.Flags().String("password", "", "Password for your Single Sign On")
	logonCmd.Flags().String("url", "", "URL for your Single Sign On")

	logonCmd.Flags().StringP("profile", "p", "default", "AWS profile to store temp credentials against")
	logonCmd.Flags().StringP("role", "r", "", "Role to auto select, if one one role available it is auto selected")

	logonCmd.Flags().Bool("token", false, "If set the token is displayed (useful for tools that don't use aws credentials file)")
	logonCmd.Flags().Int64("durationSeconds", 3600, "Use to define the duration of session validity in seconds")

	bindFlags(RootCmd, "logon")
}

func execLogon(cmd *cobra.Command, args []string) {

	uri := viper.GetString("url")

	if len(uri) == 0 {
		fatalExit(fmt.Errorf("Missing config `--url` must be specified"))
	}
	fileMustExist(filepath.Join(homeDirectory(), ".aws", "credentials"))

	u, p, err := Credentials(viper.GetString("username"), viper.GetString("password"))
	fatalExit(err, "Getting user credentials")
	sso := SSO{
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
		URL: uri,
	}

	saml, err := sso.SingleSignOn(u, p)
	fatalExit(err, "Single sign on")

	cache, close := loadCache()
	defer close()

	durationSeconds := viper.GetInt64("durationSeconds")

	arns, err := ExtractRoles(saml, cache, durationSeconds)
	fatalExit(err, "Extracting roles")
	arn, err := SelectRole(viper.GetString("role"), arns)
	fatalExit(err, "Role selection")

	if !arn.hasCredentials() {
		assumeRole(session.New(), arn, saml.AsAssertion(), durationSeconds)
	}
	if viper.GetBool("token") {
		fmt.Printf("AWS_SESSION_TOKEN=%s\n", arn.sessionToken)
	}

	err = UpdateAwsConfigFile(viper.GetString("profile"), arn.accessKeyID, arn.secretAccessKey, arn.sessionToken)
	fatalExit(err, "Updating local .aws/credentials")
}

func loadCache() (*AccountAliasCache, func()) {
	cache := &AccountAliasCache{}
	cachePath := filepath.Join(homeDirectory(), configPath, "cache")
	fileMustExist(cachePath)
	f, err := os.Open(cachePath)
	if err != nil {
		fatalExit(err)
	}
	cache.Read(f)
	return cache, func() {
		f.Close()
		f, err := os.OpenFile(cachePath, os.O_RDWR, 0)
		if err != nil {
			journal("Unable to write cache: %v : '%s'", err, cachePath)
		}
		cache.Write(f)
		err = f.Close()
		if err != nil {
			journal("Unable to write cache: %v : '%s'", err, cachePath)
		}
	}

}
