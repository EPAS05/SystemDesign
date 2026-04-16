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
	ErrEnum              = errors.New("enums can only be created under the Enumerations root (ID=3)")
	ErrCantDeleteEnum    = errors.New("use DeleteEnum to delete enumeration nodes")
	ErrCustomerNotFound  = errors.New("customer not found")
	ErrInvoiceNotDraft   = errors.New("only draft invoices can be modified")
)
