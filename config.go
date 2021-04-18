package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mibzman/gmitohtml/pkg/gmitohtml"
	"gopkg.in/yaml.v3"
)

type certConfig struct {
	Cert string
	Key  string

	cert tls.Certificate
}

type appConfig struct {
	Bookmarks map[string]string

	Certs map[string]*certConfig
}

var config = &appConfig{
	Bookmarks: make(map[string]string),

	Certs: make(map[string]*certConfig),
}

func defaultConfigPath() string {
	homedir, err := os.UserHomeDir()
	if err == nil && homedir != "" {
		return path.Join(homedir, ".config", "gmitohtml", "config.yaml")
	}
	return ""
}

func readconfig(configPath string) error {
	if configPath == "" {
		return errors.New("file unspecified")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	var newConfig *appConfig
	err = yaml.Unmarshal(configData, &newConfig)
	if err != nil {
		return err
	}
	config = newConfig

	return nil
}

func saveConfig(configPath string) error {
	config.Bookmarks = gmitohtml.GetBookmarks()

	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %s", err)
	}

	os.MkdirAll(path.Dir(configPath), 0755) // Ignore error

	err = ioutil.WriteFile(configPath, out, 0644)
	if err != nil {
		return fmt.Errorf("failed to save configuration to %s: %s", configPath, err)
	}
	return nil
}
