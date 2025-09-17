package openmeteo

import "github.com/odit-bit/jagatai/jagat/agent/tooldef"

const (
	Namespace = "openmeteo"
)

func init() {
	tooldef.Register(Namespace, NewTooldef)
}
