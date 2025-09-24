# Jagat

Jagat is ready to use AI agents server. It is designed to be stateless and extensible, allowing for the integration of various language models and custom tools.

## Core Features

- **Stateless by Design**: The agent does not manage conversation history; the client is responsible for sending the full message history with each request.
- **Pluggable LLM Providers**: Supports Ollama, Google Gemini. The provider can be configured in the `config.yaml` file.
- **Extensible Tool System**: New functionalities can be added by creating custom tools. The agent uses a registration mechanism to discover and integrate these tools at runtime.
- **Graph-Based Execution(in-progress)**: Manages the conversation flow and tool usage through a graph of executable nodes, allowing for more complex and controlled interactions.
- **Flexible Configuration**: Configuration is handled through a `config.yaml` file, with overrides possible via environment variables and command-line flags.

## How It Works

The system is composed of a central agent that orchestrates the workflow.

1.  A request is received by the REST API, which includes the entire conversation history.
2.  The `Agent` processes the request using a graph-based execution model (`completionDag`).
3.  The `AgentNode` determines the next step, which could be a direct response or a tool call, by communicating with the configured LLM provider.
4.  If a tool is required, the request is routed to the appropriate `ToolNode`, which executes the tool's logic and returns the result to the agent.
5.  The agent processes the tool's output and generates a final response.

## Getting Started

### 1\. Configuration

The application is configured using a `config.yaml` file. A default configuration is provided, which you can customize.

**Example `config.yaml`:**

```yaml
server:
  address: "127.0.0.1:11823"
  debug: false

provider: # llm backend
  name: "ollama"
  model: "qwen3:1.7b"
  apikey: "" # Set via JAGATAI_PROVIDER_APIKEY env var

tools:
  - name: "clock"
  - name: "osm"
```

### 2\. Running the Server

Use the `run` command to start the HTTP server.

```shell
go run ./main.go run --config path/to/your/config.yaml
```

or build

```shell
go build -o ./bin/jagatai ./main.go
```

### 3\. Making a Request

Interact with the agent through the REST API.

```shell
curl --request POST \
  --url http://localhost:11823/v1/chat/completions \
  --header 'Content-Type: application/json' \
  --data '{
    "content": [
        {
            "role": "user",
            "parts": [
                {
                    "text": "what time is it?"
                }
            ]
        }
    ]
}'
```

## Extending the Agent

To add a new tool:

1.  Implement the `agent.ToolProvider` interface.
2.  Register your new tool using `tooldef.Register` in an `init()` function.
3.  Import your tool's package into the main application using a blank import (`_ "path/to/your/tool"`).
4.  Add the tool's configuration to your `config.yaml` file.

see more at `document`
