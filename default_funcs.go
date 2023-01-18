package myrddin

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func (m *Myrddin) Reader(filename string) string {

	file, err := m.store.Open("/" + filename)
	if err != nil {
		err = fmt.Errorf("open of `%s` failed with: %w", filename, err)
		panic(err)
	}

	ret, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("read of `%s` failed with: %w", filename, err)
		panic(err)
	}
	return string(ret)
}

func (m *Myrddin) Loader(method string, filename string) string {
	file, err := m.store.Open("/" + filename)
	if err != nil {
		err = fmt.Errorf("%s load of `%s` failed with: %w", method, filename, err)
		panic(err)
	}
	defer file.Close()

	buf := new(strings.Builder)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	_, err = io.Copy(enc, file)
	if err != nil {
		err = fmt.Errorf("%s read of %s failed with: %w", method, filename, err)
		panic(err)
	}
	return buf.String()
}

func (m *Myrddin) PngLoader(filename string) string {
	return m.Loader("png", filename)
}

func (m *Myrddin) JsonLoader(filename string) string {
	return m.Loader("json", filename)
}

func (m *Myrddin) SvgLoader(filename string) string {
	return m.Loader("svg", filename)
}

func DefaultLoaders() Option {
	return func(m *Myrddin) error {
		m.funcMap["read"] = m.Reader
		m.funcMap["png"] = m.PngLoader
		m.funcMap["json"] = m.JsonLoader
		m.funcMap["svg"] = m.SvgLoader
		return nil
	}
}
