package env

func New() *Store {
	s := &Store{}
	s.kv = make(map[string]interface{})
	return s
}
