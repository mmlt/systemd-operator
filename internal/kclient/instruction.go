package kclient

import (
	"fmt"
)

// OpCode represents an Add, Update or Delete operation on an Object in the target device.
//go:generate stringer -type=OpCode
type OpCode int

const (
	Add OpCode = iota
	Update
	Delete
	Idle
)

// Instruction detected by API Server client.
type Instruction struct {
	// OpCode represents the kind of change; Add, Update, Delete etc.
	OpCode OpCode
	// DesiredState holds the desired state of the object.
	DesiredState *Node
}

func (in *Instruction) String() string {
	return fmt.Sprint( in.OpCode.String(), " ", in.DesiredState.String())
}