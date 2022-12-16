package myrddin

import (
	"fmt"

	"github.com/taubyte/myrddin/module"
)

type Option func(m *Myrddin) error

func Function(name string, f interface{}) Option {
	return func(m *Myrddin) error {
		if _, k := m.funcMap[name]; k == true {
			return fmt.Errorf("Duplicate function key: `%s`", name)
		}
		m.funcMap[name] = f
		return nil
	}
}

func Data(name string, data interface{}) Option {
	return func(m *Myrddin) error {
		if _, k := m.data[name]; k == true {
			return fmt.Errorf("Duplicate data key: `%s`", name)
		}
		m.data[name] = data
		return nil
	}
}

func Module(mod module.Module) Option {
	return func(m *Myrddin) error {
		return m.AddModule(mod)
	}
}
