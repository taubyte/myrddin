package env

import (
	"fmt"
)

func (e *Store) Set(name string, value interface{}) error {
	e.kv[name] = value
	return nil
}

func (e *Store) Get(name string) (interface{}, error) {
	if v, ok := e.kv[name]; ok == true {
		return v, nil
	}
	return nil, fmt.Errorf("Environment variable `%s` does not exist!", name)
}

func (e *Store) Reset() {
	e.kv = make(map[string]interface{})
}
