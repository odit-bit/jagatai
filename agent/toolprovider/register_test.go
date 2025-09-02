package toolprovider_test

import (
	"testing"

	"github.com/odit-bit/jagatai/agent/tooldef"
)

func Test_tools(t *testing.T) {
	if count := tooldef.Count(); count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}
