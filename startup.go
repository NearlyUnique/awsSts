package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type (
	options struct {
		profileName  string
		targetURL    string
		username     string
		password     string
		role         string
		dumpWork     bool
		verbose      bool
		version      bool
		loop         bool
		displayToken bool
	}
)

func startupOptions() *options {
	o := options{}
	var err error

	p := flag.String("profile", "saml", "The profile to store these temproary credentials (use 'default' to make it the default)")
	host := flag.String("host", "", "URL of form used to loging to your STS server, \n\te.g. https://sts.domain.company.org/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices")
	role := flag.String("role", "", "When multiple roles are available use this one automatically")
	v := flag.Bool("version", false, "Display version")
	dx := flag.Bool("dump-work", false, "dump HTML and XML working content from AWS")
	vb := flag.Bool("verbose", false, "Verbose output")
	loop := flag.Bool("auto", false, "If auto is enabled will allow refresh key without login details")
	token := flag.Bool("token", false, "Display temporary AWS session token")
	usage := flag.Usage

	flag.Usage = func() {
		fmt.Printf("AWS STS temporary credentials helper\n"+
			"The following environment variables will be read\n"+
			"\t%s - Url for STS login\n"+
			"\t%s - Username to loing with\n"+
			"\t%s - Password to login with\n", urlEnv, userEnv, passEnv)
		usage()
	}
	flag.Parse()

	verbose = *vb

	addEnvironmentVars(&o)

	o.profileName = *p
	o.dumpWork = *dx
	o.version = *v
	o.loop = *loop
	o.role = *role
	o.displayToken = *token
	if len(*host) > 0 {
		o.targetURL = *host
	} else {
		o.targetURL, err = getLoginURL()
	}

	if len(o.targetURL) == 0 {
		exitErr(err, "Host not specified, %s environment variable or --host not specified", urlEnv)
	}

	return &o
}
func getLoginURL() (string, error) {
	targetURL := os.Getenv(urlEnv)
	if len(targetURL) == 0 {
		return "", errors.New("Missing URL")
	}
	log("Checking env..\n%s=%s\n", urlEnv, targetURL)
	return targetURL, nil
}
func addEnvironmentVars(o *options) {
	o.username = os.Getenv(userEnv)
	o.password = os.Getenv(passEnv)
}
