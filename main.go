package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

const (
	urlEnv      = "AWSSTS_URL"
	userEnv     = "AWSSTS_USER"
	passEnv     = "AWSSTS_PASS"
	keyUsername = "UserName"
	keyPassword = "Password"
)

type (
	// AttributeValue .
	AttributeValue struct {
		Name  string   `xml:"Name,attr"`
		Value []string `xml:"AttributeValue"`
	}
	// Response .
	Response struct {
		XMLName   xml.Name         `xml:"Response"`
		Assertion []AttributeValue `xml:"Assertion>AttributeStatement>Attribute"`
	}
	//Arn principal and role
	Arn struct {
		principal, role string
	}
)

func main() {
	options := startupOptions()

	if options.version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	resp := webPageLogin(options)
	arn, assertion := getAwsTempToken(options.role, resp)

	while(func() bool {
		updateAwsConfig(options, arn, assertion)
		return options.loop && anyKeyOrQuit()
	})
}
