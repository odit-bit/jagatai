package driver

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/tooldef"
	"github.com/odit-bit/jagatai/agent/toolprovider/xtime"
	"github.com/stretchr/testify/assert"
)

func Test_genai(t *testing.T) {
	tool := xtime.NewTooldef(tooldef.Config{}).Tooling()

	fd := ToFunctionDeclaration(&tool)

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

func Test_message_convertUser(t *testing.T) {
	argsMap := map[string]any{
		"lat":  10,
		"long": 20,
	}
	b, _ := json.Marshal(argsMap)

	tc := agent.ToolCall{
		Function: agent.FunctionCall{
			Name:      "test",
			Arguments: string(b),
		},
	}

	act := &agent.Message{
		Text: "message text",
		Data: &agent.Blob{
			Bytes: []byte("this image place holder"),
			Mime:  "image/jpeg",
		},
		Toolcalls: []agent.ToolCall{tc},
		ToolResponse: &agent.ToolResponse{
			Output: map[string]any{"result": "tool response"},
		},
	}

	c, err := convertUser(act)

	if err != nil {
		t.Fatal(err)
	}

	if c.Parts == nil {
		t.Fatalf("parts cannot be nil")
	}

	for _, part := range c.Parts {
		if fc := part.FunctionResponse; fc != nil {
			assert.Equal(t, act.ToolResponse.Output, fc.Response)
		}
		if text := part.Text; text != "" {
			assert.Equal(t, act.Text, text)
		}
		if image := part.InlineData; image != nil {
			assert.Equal(t, act.Data.Bytes, image.Data)
			assert.Equal(t, act.Data.Mime, image.MIMEType)
		}
	}
}
