package ast

import "github.com/pkg/errors"

var (
	ErrInvalidFilePath  = errors.New("invalid file path")
	ErrInvalidEmptyBody = errors.New("invalid file empty body")
)
