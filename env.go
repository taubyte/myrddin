package myrddin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v3"

	"text/template"
)

type Environment struct {
	*Myrddin
	funcMap template.FuncMap
	data    map[string]interface{}
}

type EnvVariable struct {
	Name  string
	Value interface{}
}

type EnvironmentFromYaml map[string]interface{}

func (e *Environment) processEnvironmentTemplate() (io.Reader, error) {
	var env_yaml_data []byte

	env_yaml, err := e.store.Open(EnvironmentFileName)
	if err == nil {
		defer env_yaml.Close()
		env_yaml_data, err = ioutil.ReadAll(env_yaml)
		if err != nil {
			return nil, fmt.Errorf("Reading template file %s, failed with: %w", EnvironmentFileName, err)
		}
	} else {
		env_yaml_data = []byte{}
	}

	_template, err := template.New("Env").Funcs(e.funcMap).Parse(string(env_yaml_data))
	if err != nil {
		return nil, fmt.Errorf("Parsing template file %s, failed with: %w", EnvironmentFileName, err)
	}

	var buf bytes.Buffer

	err = _template.Execute(&buf, e.data)
	if err != nil {
		return nil, fmt.Errorf("Executing template file %s, failed with: %w", EnvironmentFileName, err)
	}

	return &buf, nil
}

func (e *Environment) parseEnvironment() error {
	yamlFile, err := e.processEnvironmentTemplate()
	if err != nil {
		return err
	}

	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return err
	}

	_env := make(EnvironmentFromYaml)

	err = yaml.Unmarshal(byteValue, &_env)
	if err != nil {
		return err
	}

	for k, v := range _env {
		err = e.set(EnvVariable{Name: k, Value: v})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Myrddin) Environment() *Environment {
	e := &Environment{Myrddin: m}

	e.funcMap = template.FuncMap{
		"env": func(n string) string { return os.Getenv(n) },
	}

	e.data = map[string]interface{}{
		"version":  Version,
		"hostname": func() string { h, _ := os.Hostname(); return h }(),
	}

	return e
}

func (e *Environment) set(vars ...EnvVariable) error {
	var err error = nil
	for _, v := range vars {
		err = e.Myrddin.env.Set(v.Name, v.Value)
		if err != nil {
			break
		}
	}
	return err
}

func (e *Environment) reset() {
	e.Myrddin.env.Reset()
}
