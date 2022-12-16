package myrddin

import (
	"text/template"

	"github.com/spf13/afero"
	"github.com/taubyte/myrddin/env"
)

type Myrddin struct {
	store afero.Fs
	env   *env.Store

	config interface{}

	funcMap template.FuncMap
	data    map[string]interface{}
}
