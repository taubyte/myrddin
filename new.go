package myrddin

import (
	"os"
	"text/template"

	"github.com/taubyte/myrddin/env"
)

func New(tgt interface{}, options ...Option) (*Myrddin, error) {
	m := &Myrddin{
		env: env.New(),
	}

	m.funcMap = template.FuncMap{
		"hostname": func() string { h, _ := os.Hostname(); return h },
		"env":      func(name string) interface{} { v, _ := m.env.Get(name); return v },
	}

	m.data = map[string]interface{}{
		"version": Version,
	}

	m.config = tgt

	for _, opt := range options {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
