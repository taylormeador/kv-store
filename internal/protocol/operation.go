package protocol

import (
	"errors"
)

var ErrInvalidOperation error = errors.New("invalid operation")

type Operation struct {
	Directive Directive
	Key       string
	Value     string
}

// Define Directive type using Enum pattern
type Directive string

const (
	GetDirective    Directive = "GET"
	SetDirective    Directive = "SET"
	DeleteDirective Directive = "DELETE"
	ExistsDirective Directive = "EXISTS"
	TxDirective     Directive = "TX"
)

func (d Directive) IsValid() bool {
	switch d {
	case GetDirective, SetDirective, DeleteDirective, ExistsDirective, TxDirective:
		return true
	default:
		return false
	}
}

// This is not exhaustive, it is to be used in conjunction with parseNextOperation()
func (o *Operation) IsValid() bool {
	if o.Key == "" {
		return false
	}

	if o.Directive == SetDirective {
		if o.Value == "" {
			return false
		}
	} else {
		if o.Value != "" {
			return false
		}
	}

	return true
}

// Returns a string of a valid Operation
func (op *Operation) String() string {
	result := string(op.Directive) + " " + op.Key
	if op.Directive == SetDirective {
		result += " " + op.Value
	}
	return result
}
