package agent

type options struct {
	tools       Tools
	toolMaxCall int
}

type OptionFunc func(o *options)

// bind tool into model
func WithTool(tools ...ToolProvider) OptionFunc {
	return func(o *options) {
		o.tools = tools
	}
}

// set maximum tool call agent can invoke, not implemented yet
func WithMaxToolCall(n int) OptionFunc {
	return func(o *options) {
		o.toolMaxCall = n
	}
}
