### Core Concepts 🧠

The tool system is built around a few key interfaces and structs defined in `jagat/agent/tool.go`:

- **`ToolProvider`**: This is the main interface that every tool must implement. It combines two other interfaces: `ToolDefinition` and `XTool`.
- **`ToolDefinition`**: This interface is responsible for defining the tool's schema, which is sent to the LLM. It includes the tool's name, description, and the parameters it accepts.
- **`XTool`**: This interface defines the tool's runtime behavior. It has a `Call` method that executes the tool's logic and a `Ping` method to check if the tool is ready to be used.

---

### Tool Registration ✍️

Tools are dynamically registered and loaded using the `tooldef` package. This makes it possible to add or remove tools without changing the core agent logic.

- **Registration**: Each tool provider registers itself using the `tooldef.Register` function, which is usually done in the tool's `init` function. This function maps a unique string name to the tool's constructor.
- **Building Tools**: The `tooldef.Build` function is called at application startup. It iterates through the tools defined in the configuration, finds their corresponding registered constructor, and initializes them.

---

### Tool Implementation Example: `xtime` ⏰

The `xtime` tool is a great example of a simple tool implementation. It provides the current time to the agent. Here's a breakdown of its key components in `jagat/agent/toolprovider/xtime/time.go`:

- **Definition**: The tool's definition specifies its name (`get_current_time`), a description for the LLM, and its parameters.
- **`Call` Method**: This is where the core logic of the tool resides. For `xtime`, it simply gets the current time and returns it in a `ToolResponse`.
- **Registration**: The tool registers itself with the name "clock" in its `init` function.

---

### Configuration ⚙️

Tools are configured in the `tools` section of the `config.yaml` file. Each entry in this list specifies the name of the tool and any additional configuration it might need, such as an API key or an endpoint.

**Example `config.yaml`:**

```yaml
tools:
  - name: "clock"
  - name: "osm"
  - name: "openmeteo"
    endpoint: "https://api.open-meteo.com"
  - name: "tavily"
    endpoint: "https://api.tavily.com"
    apikey: "YOUR_API_KEY"
```
