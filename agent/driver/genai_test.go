package driver_test

import (
	"strings"
	"testing"

	"github.com/odit-bit/jagatai/agent/driver"
	"github.com/odit-bit/jagatai/agent/tooldef"
	"github.com/odit-bit/jagatai/agent/toolprovider/xtime"
)

func TestXxx(t *testing.T) {
	tool := xtime.NewTooldef(tooldef.Config{}).Tooling()

	fd := driver.ToFunctionDeclaration(&tool)

	//function
	if fd.Description != tool.Function.Description {
		t.Fatalf("different function description, got %v expect %v", tool.Function.Description, fd.Description)
	}

	//parameter
	actP := fd.Parameters
	expP := tool.Function.Parameters
	for k, v := range actP.Properties {
		vExpect := expP.Properties[k]
		if vExpect.Description != v.Description {
			t.Fatalf("different parameter description, got %v expect %v", vExpect.Description, v.Description)
		}

		if vExpect.Type != strings.ToLower(string(v.Type)) {
			t.Fatalf("different type , got %v expect %v", vExpect.Type, v.Type)

		}
	}
}
