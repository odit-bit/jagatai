package tooldef

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/odit-bit/jagatai/jagat/agent"
)

type MCPConfig struct {
	Endpoint string
}

type MCPRegistry struct {
	session *mcp.ClientSession
}

func NewMCPRegistry(ctx context.Context, conf MCPConfig) (MCPRegistry, error) {
	c := mcp.NewClient(&mcp.Implementation{
		Name:    "jagat-mcp-client",
		Version: "0.0.1",
	}, &mcp.ClientOptions{})

	session, err := c.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: conf.Endpoint}, &mcp.ClientSessionOptions{})
	if err != nil {
		return MCPRegistry{}, fmt.Errorf("failed to connect mcp server: %v", err)
	}

	return MCPRegistry{
		session: session,
	}, nil
}

func (registry *MCPRegistry) Close() error {
	return registry.session.Close()
}

func (registry *MCPRegistry) ListTools(ctx context.Context, tp []agent.ToolProvider) ([]agent.ToolProvider, error) {
	log.Printf("Connected to server (session ID: %s)", registry.session.ID())
	// First, list available tools.
	log.Println("Listing available tools at MCP Server...")
	tools, err := registry.session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	for _, tool := range tools.Tools {

		log.Printf("  - %s: %s\n", tool.Name, tool.Description)

		b, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("mcp registry failed to read tool input schema %s: %v", tool.Name, err)
		}

		var param agent.ParameterSchema
		json.Unmarshal(b, &param)

		t := McpToolProvider{
			session:     registry.session,
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: param,
		}

		tp = append(tp, &t)
	}

	return tp, nil
}

var _ agent.ToolProvider = (*McpToolProvider)(nil)

type McpToolProvider struct {
	session     *mcp.ClientSession
	Name        string
	Description string
	InputSchema agent.ParameterSchema
}

func (mcptp *McpToolProvider) Call(ctx context.Context, fn agent.FunctionCall) (*agent.ToolResponse, error) {

	param := map[string]string{}
	if err := json.Unmarshal([]byte(fn.Arguments), &param); err != nil {
		slog.Debug(err.Error())
		var tr agent.ToolResponse
		tr.Output["error"] = "wrong format arguments"
		return &tr, nil
	}

	res, err := mcptp.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      fn.Name,
		Arguments: param,
	})

	if err != nil {
		return nil, fmt.Errorf("mcp client failed calling tool %s: %w", fn.Name, err)
	}

	log.Printf("DEBUG MCP tool_call response: %s \n", res.StructuredContent)
	return &agent.ToolResponse{
		Name:   fn.Name,
		Output: res.StructuredContent.(map[string]any),
	}, nil
}

func (mcptp *McpToolProvider) Ping(ctx context.Context) error {
	return nil
}

func (mcptp *McpToolProvider) Def() agent.Tool {

	// log.Printf("  - input_schema: %s\n", string(b))

	return agent.Tool{
		Type: "function",
		Function: agent.Function{
			Name:        mcptp.Name,
			Description: mcptp.Description,
			Parameters:  mcptp.InputSchema,
		},
	}
}
