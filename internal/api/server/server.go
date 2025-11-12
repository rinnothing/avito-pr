package server

import (
	"github.com/rinnothing/avito-pr/api/gen"
)

var _ gen.ServerInterface = &serverImplementation{}

type serverImplementation struct {
}

func New() *serverImplementation {
	return &serverImplementation{}
}
