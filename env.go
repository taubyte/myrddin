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

type Environement struct {
	*Myrddin
	funcMap template.FuncMap
	data    map[string]interface{}
}

type EnvVariable struct {
	Name  string
	Value interface{}
}

type EnvironementFromYaml map[string]interface{}

func (e *Environement) processEnvironementTemplate() (io.Reader, error) {
	var env_yaml_data []byte

	env_yaml, err := e.store.Open(EnvironementFileName)
	if err == nil {
		defer env_yaml.Close()
		env_yaml_data, err = ioutil.ReadAll(env_yaml)
		if err != nil {
			return nil, fmt.Errorf("Reading template file %s, failed with: %w", EnvironementFileName, err)
		}
	} else {
		env_yaml_data = []byte{}
	}

	tmpl, err := template.New("Env").Funcs(e.funcMap).Parse(string(env_yaml_data))
	if err != nil {
		return nil, fmt.Errorf("Parsing template file %s, failed with: %w", EnvironementFileName, err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, e.data)
	if err != nil {
		return nil, fmt.Errorf("Executing template file %s, failed with: %w", EnvironementFileName, err)
	}

	return &buf, nil
}

func (e *Environement) parseEnvironement() error {
	yfile, err := e.processEnvironementTemplate()
	if err != nil {
		return err
	}

	byteValue, err := ioutil.ReadAll(yfile)
	if err != nil {
		return err
	}

	_env := make(EnvironementFromYaml)

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

func (m *Myrddin) Environement() *Environement {
	e := &Environement{Myrddin: m}

	e.funcMap = template.FuncMap{
		"env": func(n string) string { return os.Getenv(n) },
	}

	e.data = map[string]interface{}{
		"version":  Version,
		"hostname": func() string { h, _ := os.Hostname(); return h }(),
	}

	return e
}

func (e *Environement) set(vars ...EnvVariable) error {
	var err error = nil
	for _, v := range vars {
		err = e.Myrddin.env.Set(v.Name, v.Value)
		if err != nil {
			break
		}
	}
	return err
}

func (e *Environement) reset() {
	e.Myrddin.env.Reset()
}
