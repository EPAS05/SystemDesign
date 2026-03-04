package repository

import (
	"errors"
)

var (
	ErrNotFound          = errors.New("node not found")
	ErrInvalidParent     = errors.New("invalid parent node")
	ErrTypeMismatch      = errors.New("child type not allowed for this parent")
	ErrCannotDeleteTrash = errors.New("cannot delete trash node")
	ErrCycleDetected     = errors.New("Cycle detected")
)
