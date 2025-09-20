## `config` Package Documentation

The `config` package is a vital part of the Jagat project, providing a robust and flexible way to manage application configuration. It aggregates settings from various sources, validates them, and makes them available to the rest of the application.

---

### Configuration Structure

The main configuration is defined by the `Config` struct in `jagat/config/config.go`. Here's a breakdown of the key sections:

#### `ServerConfig`

| Field     | Type   | Description                                                           |
| :-------- | :----- | :-------------------------------------------------------------------- |
| `Address` | string | The address and port the server listens on (e.g., "127.0.0.1:11823"). |
| `Debug`   | bool   | Enables or disables debug logging.                                    |

#### `Provider`

This section configures the external LLM provider.

| Field      | Type          | Description                                                   |
| :--------- | :------------ | :------------------------------------------------------------ |
| `Name`     | string        | The name of the provider (e.g., "ollama", "openai", "genai"). |
| `Model`    | string        | The specific model to use (e.g., "qwen3:1.7b").               |
| `ApiKey`   | string        | The API key for the provider.                                 |
| `Endpoint` | string        | The endpoint URL for the provider.                            |
| `Extra`    | driver.Config | Extra provider-specific settings.                             |

---

### How it works

The configuration is loaded at application startup by calling the `LoadAndValidate` function. Here’s the order of precedence for loading settings:

1.  **Default `config.yaml`**: The application starts with a default configuration.
2.  **Custom YAML File**: You can specify a custom configuration file using the `--config` flag.
3.  **Environment Variables**: Override any setting by using an environment variable with the prefix `JAGATAI_` (e.g., `JAGATAI_SERVER_ADDRESS="0.0.0.0:8080"`).
4.  **Command-Line Flags**: The highest precedence. For example, `--addr "localhost:9999"` will override the address from all other sources.

**Example `config.yaml`:**

```yaml
server:
  address: "127.0.0.1:11823"
  debug: false

provider:
  name: "ollama"
  model: "qwen3:1.7b"
  apikey: ""
  extra:
    endpoint: "http://127.0.0.1:11434"

tools:
  - name: "clock"
  - name: "osm"
```

---

### Command-Line Flags 🚩

The `flags.go` file defines a set of command-line flags for overriding the configuration.

| Flag        | Environment Variable      | Purpose                              |
| :---------- | :------------------------ | :----------------------------------- |
| `--addr`    | `JAGATAI_SERVER_ADDRESS`  | Sets the server address.             |
| `--debug`   | `JAGATAI_SERVER_DEBUG`    | Enables debug mode.                  |
| `--config`  | (none)                    | Path to a custom configuration file. |
| `--p_key`   | `JAGATAI_PROVIDER_APIKEY` | Sets the provider's API key.         |
| `--p_name`  | `JAGATAI_PROVIDER_NAME`   | Sets the provider's name.            |
| `--p_model` | `JAGATAI_PROVIDER_MODEL`  | Sets the provider's model.           |
