package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
)

func execLogon() {
	uri := viper.GetString("url")

	if len(uri) == 0 {
		fatalExit(fmt.Errorf("Missing config `--url` must be specified"))
	}
	fileMustExist(filepath.Join(homeDirectory(), ".aws", "credentials"))

	u, p, err := Credentials(viper.GetString("username"), viper.GetString("password"))
	fatalExit(err, "Getting user credentials")
	sso := SSO{
		Client: &http.Client{
			Timeout: time.Second * 60,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		URL: uri,
	}

	saml, err := sso.SingleSignOn(u, p)
	fatalExit(err, "Single sign on")

	cache, close := loadCache()
	defer close()

	arns, err := ExtractRoles(saml, cache)
	fatalExit(err, "Extracting roles")
	arn, err := SelectRole(viper.GetString("role"), arns)
	fatalExit(err, "Role selection")

	if !arn.hasCredentials() {
		s, err := session.NewSession()
		fatalExit(err, "New aws session")
		assumeRole(s, arn, saml.AsAssertion())
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
