package module

type Module interface {
	Functions() []ModuleFunction
	Data() map[string]interface{}
}

type ModuleFunction interface {
	Name() string
	Function() interface{}
}
