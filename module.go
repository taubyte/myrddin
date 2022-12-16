package myrddin

import (
	"errors"

	"github.com/taubyte/myrddin/module"
)

func (m *Myrddin) AddModule(p module.Module) error {
	if p == nil {
		return errors.New("Invalid nil module")
	}

	funcs := p.Functions()
	data := p.Data()

	if (funcs == nil || len(funcs) == 0) || (data == nil || len(data) == 0) {
		return errors.New("Invalid empty module")
	}

	// Functions
	for _, f := range funcs {
		err := Function(f.Name(), f.Function())(m)
		if err != nil {
			return err
		}
	}

	// Data
	for k, v := range data {
		err := Data(k, v)(m)
		if err != nil {
			return err
		}
	}

	return nil

}
