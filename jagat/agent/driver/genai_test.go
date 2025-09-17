package driver

import (
	"strings"
	"testing"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	"github.com/odit-bit/jagatai/jagat/agent/toolprovider/xtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
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

func Test_messageToContent(t *testing.T) {
	msgs := &agent.Message{
		Role: agent.RoleUser,
		Parts: []*agent.Part{
			{
				Text: "test",
			},
		},
	}
	content := &genai.Content{}
	err := messageToContent(msgs, content)
	require.ErrorIs(t, nil, err)

	assert.Equal(t, (*genai.Blob)(nil), content.Parts[0].InlineData)
}
