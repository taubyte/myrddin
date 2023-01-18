package myrddin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type ParseOption Option

func Define(name string, data interface{}) ParseOption {
	return func(m *Myrddin) error {
		return m.Environement().set(EnvVariable{Name: name, Value: data})
	}
}

func (m *Myrddin) Parse(options ...ParseOption) error {

	if m.store == nil {
		return errors.New("Please load data first")
	}

	m.Environement().reset()

	for _, opt := range options {
		err := opt(m)
		if err != nil {
			return fmt.Errorf("Processing options failed with %w", err)
		}
	}

	err := m.Environement().parseEnvironement()
	if err != nil {
		return err
	}

	err = m.parseAllSections()
	if err != nil {
		return fmt.Errorf("Failed calling parse all sections with err: %w", err)
	}

	return nil
}

func (m *Myrddin) readFileOS(file string) (b []byte, err error) {
	f, err := m.store.Open(file)
	if err != nil {
		return
	}
	b, err = ioutil.ReadAll(f)
	return
}

func (m *Myrddin) createTemplateEngine() (*template.Template, error) {

	base_template := template.New("Myrddin")

	templates := make([]string, 0)

	err := afero.Walk(m.store, "/", func(path string, info fs.FileInfo, err error) error {

		if info != nil && info.IsDir() == true {
			return nil
		}

		if path != EnvironementFileName && filepath.Dir(path) == "/" {
			return nil
		}

		if strings.HasSuffix(path, ".tmpl") == false && strings.HasSuffix(path, ".tpl") == false && strings.HasSuffix(path, ".yaml") == false {
			return nil
		}

		data, err := m.readFileOS(path)
		if err != nil {
			return err
		}

		_, err = base_template.New(path).Funcs(m.funcMap).Parse(string(data))
		templates = append(templates, path)
		return err
	})

	var buf bytes.Buffer
	for _, path := range templates {
		f_yaml, err := m.store.Open(path)
		if err != nil {
			return nil, fmt.Errorf("Opening file %s, failed with: %w", path, err)
		}

		f_yaml_data, err := ioutil.ReadAll(f_yaml)
		if err != nil {
			return nil, fmt.Errorf("Reading file %s, failed with: %w", path, err)
		}

		tmpl, err := base_template.New(path).Funcs(m.funcMap).Parse(string(f_yaml_data) + "\n")
		if err != nil {
			return nil, fmt.Errorf("Parsing file %s, failed with: %w", path, err)
		}

		buf.Reset()
		err = tmpl.Execute(&buf, m.data)
		if err != nil {
			return nil, fmt.Errorf("Executing file %s, failed with: %w", path, err)
		}

	}

	return base_template, err
}

func (m *Myrddin) parseAllSections() error {
	base_template, err := m.createTemplateEngine()
	if err != nil {
		return err
	}

	outputFile, err := os.OpenFile(ProcessingFileName(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("Failed calling open file with err: %w", err)
	}
	defer outputFile.Close()

	err = m.exportTemplateTo(base_template, outputFile)
	if err != nil {
		return err
	}

	outputFile.Seek(0, 0)

	yamlDec := yaml.NewDecoder(outputFile)

	err = yamlDec.Decode(m.config)
	if err != io.EOF && err != nil {
		return fmt.Errorf("Decoding yaml failed with err: %w", err)
	}

	return nil
}

func (m *Myrddin) exportTemplateTo(base_template *template.Template, outputFile *os.File) error {
	_fs := afero.NewIOFS(m.store)

	return afero.Walk(m.store, "/", func(path string, info fs.FileInfo, err error) error {
		// Ignore directories || Ignore the Myrddin environment file
		if (info != nil && info.IsDir() == true) || path == EnvironementFileName {
			return nil
		}

		// Ignore sub-directories
		if filepath.Dir(path) != "/" {
			return nil
		}

		// Ignore files without .yaml or .yml ext
		if strings.HasSuffix(path, ".yaml") == false && strings.HasSuffix(path, ".yml") == false {
			return nil
		}

		// Check and make sure file doesnt start with a /
		if path[0:1] == "/" && len(path) > 0 {
			path = path[1:]
		}

		tmpl, err := base_template.New(path).Funcs(m.funcMap).ParseFS(_fs, path)
		if err != nil {
			return fmt.Errorf("Parsing file %s, failed with: %w", path, err)
		}

		err = tmpl.Execute(outputFile, m.data)
		if err != nil {
			return fmt.Errorf("Executing file %s, failed with: %w", path, err)
		}

		// let's make sure we have a new empty line so YAML parsers do not complain
		fmt.Fprintln(outputFile)
		return nil
	})
}
