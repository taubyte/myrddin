package main

import (
	"fmt"

	"github.com/taubyte/myrddin"
	"gopkg.in/yaml.v3"
)

type MyConfig struct {
	Hostname string `yaml:"host"`
	Shell    string `yaml:"shell"`
	Name     string `yaml:"name"`
	Networks []struct {
		Name      string `yaml:"name"`
		AddrRange struct {
			Start string `yaml:"start"`
			End   string `yaml:"end"`
		} `yaml:"addr_range"`
	} `yaml:"networks"`
}

func main() {
	config := &MyConfig{}
	m, err := myrddin.New(config, myrddin.Data(
		"ver", []string{"v0.0.1"},
	))
	if err != nil {
		panic(err)
	}

	err = m.Load("config")
	if err != nil {
		panic(err)
	}

	err = m.Parse()
	if err != nil {
		panic(err)
	}

	out, _ := yaml.Marshal(config)

	fmt.Println(string(out))
}
