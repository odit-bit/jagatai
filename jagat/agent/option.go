package agent

type options struct {
	tools       Tools
	toolMaxCall int
}

type OptionFunc func(a *Agent) *Agent

// set maximum tool call agent can invoke, not implemented yet
func WithMaxToolCall(n int) OptionFunc {
	return func(a *Agent) *Agent {
		a.toolMaxCall = n
		return a
	}
}

func WithTools(tools ...ToolProvider) OptionFunc {
	return func(a *Agent) *Agent {
		a.tools = tools
		return a
	}
}
