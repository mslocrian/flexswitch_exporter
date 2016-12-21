package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

var (
	DefaultModule = Module{
		Port:  8080,
		Proto: "http",
	}
	DefaultAuth = Auth{
		Username: "",
		Password: "",
	}
)

type Config map[string]*Module

type Module struct {
	Port  int    `yaml:"port"`
	Proto string `yaml:"proto"`
	Auth  *Auth  `yaml:"auth"`

	XXX map[string]interface{} `yaml:",inline"`
}

func (c *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultModule
	type plain Module
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if err := checkOverflow(c.XXX, "module"); err != nil {
		return err
	}
	if c.Auth == nil {
		c.Auth = &DefaultAuth
	}

	return nil
}

type Auth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`

	XXX map[string]interface{} `yaml:",inline"`
}

func (c *Auth) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultAuth
	type plain Auth
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if err := checkOverflow(c.XXX, "module"); err != nil {
		return err
	}
	return nil
}

func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields in %s: %s", ctx, strings.Join(keys, ", "))
	}
	return nil
}
