![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/odit-bit/jagatai?utm_source=oss&utm_medium=github&utm_campaign=odit-bit%2Fjagatai&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)

# Jagatai Agent Server Documentation

## Overview

This document explains the structure and usage of the Agent Server, which provides a REST API for interacting with a language model agent. The server supports tools, and multiple provider (e.g., Ollama, OpenAI).

## How to Use

### 1\. Install Dependencies

To get started, download all the necessary Go modules.

```sh
go mod download
```

---

### 2\. Run the Server

Execute the `main.go` file to start the server. You can specify the address, an API key, and a tools configuration file using command-line flags. see `app/server/config/config.yaml` for more configuration.

**run command:**

```sh
go run ./cmd/server/main.go run my-model --debug
```

**build:**

```sh
go build -o ./bin/server ./cmd/server/
```

```
./bin/server --help
```

```
./bin/server run <model-name> --debug
```

### 3\. Test Endpoints

Use `curl` or any other API client to test the server's endpoints.

**Chat Completions:**

```sh
curl -X POST http://127.0.0.1:11823/v1/chat/completions \
 -H "Content-Type: application/json" \
 -d '{
 "messages": [
     {"role": "user", "content": "What is the weather like today?"}
    ]
  }'
```

## Architecture Diagram

```
+---------------------+
|   Client (User)     |
+---------------------+
          |
          | (HTTP)
          v
+---------------------+
|  Server (Jagatai)   |
|  - Echo Router      |
|  - Agent Logic      |
|  - LLM Drivers      |
|  - Tools Management |
+---------------------+
```

---

## Roadmap

Next iteration:

- Add new tools by implementing the `tooldef.Provider` interface.
- Extend LLM drivers to support other APIs.
- Add more detailed metrics for CPU, disk usage, etc.

---

## License

Licensed under the MIT License. See the `LICENSE` file for details.

---

## References

- [Ollama API Documentation](https://github.com/ollama/ollama/tree/main/docs)
- [OpenTelemetry Go](https://github.com/ollama/ollama/tree/main/docs)
- [Prometheus Client Go](https://github.com/prometheus/client_golang)
- [OpenAI](https://platform.openai.com/docs/api-reference/chat/create)

<!-- end list -->
