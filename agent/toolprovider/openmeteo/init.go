package openmeteo

import "github.com/odit-bit/jagatai/agent/tooldef"

const (
	Namespace = "openmeteo"
)

func init() {
	tooldef.Register(Namespace, NewTooldef)
}
