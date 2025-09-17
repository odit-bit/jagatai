package agent

type options struct {
	tools       ToolsMap
	toolMaxCall int
}

type OptionFunc func(o *options)

func WithTool(tools ...Tool) OptionFunc {
	return func(o *options) {
		tmap := make(ToolsMap)
		for _, t := range tools {
			tmap[t.Function.Name] = t
		}
		o.tools = tmap
	}
}

func WithMaxToolCall(n int) OptionFunc {
	return func(o *options) {
		o.toolMaxCall = n
	}
}
