package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type certConfig struct {
	Cert string
	Key  string

	cert tls.Certificate
}

type appConfig struct {
	Certs map[string]*certConfig
}

var config = &appConfig{
	Certs: make(map[string]*certConfig),
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
