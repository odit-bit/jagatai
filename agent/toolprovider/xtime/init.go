package xtime

import "github.com/odit-bit/jagatai/agent/tooldef"

const (
	Namespace = "clock"
)

func init() {
	tooldef.Register(Namespace, NewTooldef)
}
