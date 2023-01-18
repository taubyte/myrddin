package myrddin

const (
	EnvironmentFileName = "/env.yaml"
)

var (
	SpecialFiles = []string{EnvironmentFileName}
)

var ProcessingFileName = func() string {
	return "config.debug"
}
