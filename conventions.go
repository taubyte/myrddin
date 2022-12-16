package myrddin

const (
	EnvironementFileName = "/env.yaml"
)

var (
	SpecialFiles = []string{EnvironementFileName}
)

var ProcessingFileName = func() string {
	return "config.debug"
}
