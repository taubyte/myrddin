package myrddin

import (
	"io"
)

type Plugin interface {
	schema() string
	Open(uri string) (io.Reader, error)
}
