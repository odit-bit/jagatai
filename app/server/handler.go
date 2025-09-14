package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/odit-bit/jagatai/agent"
)

type ChatResponse struct {
	// Model      string        `json:"model"`
	DoneReason string        `json:"done_reason"`
	CreatedAt  time.Time     `json:"created_at"`
	Message    agent.Message `json:"message"`
}
type ChatRequest struct {
	Messages []agent.Message
	Think    bool
}

func (cr *ChatRequest) validate() error {
	if len(cr.Messages) == 0 {
		return fmt.Errorf("messages cannot be nil")
	}
	return nil
}

type Agent interface {
	Completions(ctx context.Context, req *agent.CCReq) (*agent.CCRes, error)
}

func HandleAgent(ctx context.Context, a Agent, e *echo.Echo) {
	if e == nil {
		panic("got nil parameter")
	}

	e.POST("/v1/chat/completions", func(c echo.Context) error {
		slog.Debug("got request")
		if ok := IsJsonContentType(c.Request()); !ok {
			return c.JSON(400, echo.Map{"error": "expecting json body"})
		}

		var input ChatRequest
		if err := c.Bind(&input); err != nil {
			slog.Error("failed binding", "error", err, "type", input)
			return c.JSON(400, echo.Map{"error": "bad json format"})
		}
		if err := input.validate(); err != nil {
			slog.Error("validater error", "error", err)
			return c.JSON(400, echo.Map{"error": "bad json format."})
		}

		output, err := a.Completions(c.Request().Context(), &agent.CCReq{
			Messages: input.Messages,
			Think:    input.Think,
			Stream:   false,
		})

		if err != nil {
			slog.Error("failed completion", "error", err)
			return c.JSON(400, echo.Map{"error": "server unavailable"})
		}

		cr := ChatResponse{
			DoneReason: output.Choices[0].FinishReason,
			CreatedAt:  output.Created,
			Message:    output.Choices[0].Message,
		}
		slog.Debug("request finish")
		return c.JSON(200, cr)
	})

}

func IsJsonContentType(req *http.Request) bool {
	ct := req.Header.Get("Content-Type")
	return ct == "application/json"
}

// Function to update the RAM usage metric
func updateRAMUsage() (metric.Registration, error) {
	meter := otel.Meter("jagatAI_rest_server_meter")
	ramUsage, err := meter.Int64ObservableGauge(
		"jagatAI_ram_usage_bytes",
		metric.WithDescription("Ram usage of the app in bytes"),
	)
	if err != nil {
		return nil, err
	}

	return meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		var stats runtime.MemStats

		// Get the memory statistics
		runtime.ReadMemStats(&stats)

		// Update the RAM usage metric
		o.ObserveInt64(ramUsage, int64(stats.Sys))

		// // Wait for a short period before updating again
		// time.Sleep(10 * time.Second)
		return nil
	}, ramUsage)
}
