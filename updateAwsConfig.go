package main

import (
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

func updateAwsConfig(options *options, arn *Arn, assertion string) {
	err := getTempConfigValues(options, *arn, assertion)
	exitErr(err, "Failed to update config")
}
func getTempConfigValues(options *options, arn Arn, assertion string) error {
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "AWS session")
	}
	svc := sts.New(sess)
	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(arn.principal),
		RoleArn:       aws.String(arn.role),
		SAMLAssertion: aws.String(assertion),
	}
	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to assume role:%s", arn.role))
	}
	return updateAwsConfigFile(
		options,
		*resp.Credentials.AccessKeyId,
		*resp.Credentials.SecretAccessKey,
		*resp.Credentials.SessionToken)
}
func updateAwsConfigFile(options *options, key, secret, session string) error {
	home := "~"
	for _, env := range []string{"HOME", "USERPROFILE", "HOMEPATH"} {
		home = os.Getenv(env)
		if len(home) > 0 {
			break
		}
	}
	iniFile := path.Join(home, ".aws", "credentials")

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return errors.Wrapf(err, "Loading credentials file %s", iniFile)
	}
	s, err := cfg.GetSection(options.profileName)
	if err != nil {
		s, err = cfg.NewSection(options.profileName)
	}
	if err != nil {
		return errors.Wrapf(err, "Unable to add or create section '%s' in  %s", options.profileName, iniFile)
	}
	s.Key("aws_access_key_id").SetValue(key)
	s.Key("aws_secret_access_key").SetValue(secret)
	s.Key("aws_session_token").SetValue(session)
	err = cfg.SaveTo(iniFile)
	if err != nil {
		return errors.Wrapf(err, "Saving credentials file %s", iniFile)
	}

	log("Profile %q updated\n", options.profileName)
	return nil
}
