### 1\. Create Your Custom Tool 🛠️

First, you need to create a new Go package for your custom tool. Within this package, you'll implement the `agent.ToolProvider` interface.

**`jagat/agent/tool.go`**

```go
type ToolProvider interface {
	ToolDefinition
	XTool
}

type ToolDefinition interface {
	Def() Tool
}

type XTool interface {
	Call(ctx context.Context, fn FunctionCall) (*ToolResponse, error)
	Ping(ctx context.Context) error
}
```

---

### 2\. Integrate Your Tool into the Application 🔌

For your tool to be discoverable by the Jagat application, you need to import your new tool package in your application's main entry point (e.g., in `main.go`). This is done using a blank import, which triggers the `init` function in your tool package, registering it with the system.

**`your/app/main.go`**

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	// Import your custom tool package with a blank import.
	// This is the crucial step to ensure your tool's init() function is called.
	_ "path/to/your/tools/mytool"
)

func main() {
	/*....*/
}

```

This is the same mechanism that the Jagat project itself uses in `jagat/agent/toolprovider/init_gen.go` to load its built-in tools.

---

### 3\. Configure Your Tool ⚙️

Finally, you need to configure your tool in the `config.yaml` file. This is where you'll provide the tool's name (the one you used in `tooldef.Register`) and any specific settings like `endpoint` or `apikey`.

**`config.yaml`**

```yaml
tools:
  - name: "clock" # An existing tool
  - name: "my-tool" # Your new tool
    endpoint: "https://api.my-tool.com"
    apikey: "${MY_TOOL_API_KEY}" # You can use env variables for secrets.
```

The Jagat configuration system will automatically read this section and pass the relevant `tooldef.Config` to your tool's `New` function when the application starts. You can then set any necessary environment variables (like `MY_TOOL_API_KEY`) in your deployment environment.

### How It Works: The Precedence Logic ⚙️

This implementation ensures a clear and secure way to manage your tool's API key:

1.  **Highest Precedence (config.yaml)**: Jagat's config loader (`viper`) first populates the `tooldef.Config` struct. If `apikey` is explicitly set in your `config.yaml` for `my-tool`, its value will be used directly. This is useful for development or specific setups.

2.  **Error on Failure**: If the API key cannot be found in either location, the constructor returns an error. This is crucial because it prevents the tool from being loaded in a non-functional state. The `tooldef.Build` function will catch this error and skip adding the tool, logging a warning so you know exactly what happened.

### 4\. Run

If secret not in config, set the env.

```GO
export MY_TOOL_API_KEY="my-awesome-key"
export JAGAT_PROVIDER_APIKEY="my-awesome-llm-key"
```

then run the main, it will populate all value from config file, env and flags.

```GO
go run ./main.go --config="to your config path" --debug
```

### 5\. try it.

```
curl --request POST \
  --url http://localhost:11823/v1/chat/completions \
  --header 'Authorization: Basic Og==' \
  --header 'Content-Type: application/json' \
  --data '{
	"content": [
		{
			"role": "user",
			"parts": [
				{
					"text": "what you can do ?"
				}

			]
		}
	]
}
```

or use minimal cli.

```shell
# From project root
go run ./main.go chat
```
