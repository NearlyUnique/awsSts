package cmd

import (
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

//UpdateAwsConfigFile in .aws home folder
func UpdateAwsConfigFile(profileName, id, secret, session string) error {
	cfg, path, err := readConfig("credentials")
	if err != nil {
		return err
	}

	s, err := cfg.GetSection(profileName)
	if err != nil {
		s, err = cfg.NewSection(profileName)
	}
	if err != nil {
		return errors.Wrapf(err, "Unable to add or create section '%s' in  %s", profileName, path)
	}
	s.Key("aws_access_key_id").SetValue(id)
	s.Key("aws_secret_access_key").SetValue(secret)
	s.Key("aws_session_token").SetValue(session)
	err = cfg.SaveTo(path)
	if err != nil {
		return errors.Wrapf(err, "Saving credentials file %s", path)
	}

	journal("Profile %q updated\n", profileName)
	return nil
}

func readConfig(name string) (*ini.File, string, error) {
	iniFile := filepath.Join(homeDirectory(), ".aws", name)

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return nil, iniFile, errors.Wrapf(err, "Loading %s file %s", name, iniFile)
	}
	return cfg, iniFile, nil
}
