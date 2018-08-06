package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	TemplateUrl *string                      `yaml:"TemplateUrl"`
	Region      *string                      `yaml:"Region"`
	StackName   *string                      `yaml:"StackName"`
	Parameters  *map[interface{}]interface{} `yaml:"Parameters"`
}

var (
	defaultRegion = "us-east-1"
)

func loadConfig(path string) (*Config, error) {
	apath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fInfo, err := os.Stat(apath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if fInfo.IsDir() {
		return nil, fmt.Errorf("%s is a directory", apath)
	}

	buf, err := ioutil.ReadFile(apath)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, err
	}

	if c.TemplateUrl == nil || len(*c.TemplateUrl) < 1 {
		return nil, fmt.Errorf("%s is not specify TemplateUrl", apath)
	}
	if c.StackName == nil || len(*c.StackName) < 1 {
		return nil, fmt.Errorf("%s is not specify StackName", apath)
	}
	if c.Region == nil || len(*c.Region) < 1 {
		c.Region = &defaultRegion
	}

	return &c, nil
}
