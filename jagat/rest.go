package jagat

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/odit-bit/jagatai/jagat/agent"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Request
type ChatRequest struct {
	Content []*agent.Message `json:"content"`
}

// Response
type ChatResponse struct {
	Created time.Time `json:"created"`
	Text    string    `json:"text"`
}

func (cr *ChatRequest) validate() error {
	if len(cr.Content) == 0 {
		return fmt.Errorf("messages cannot be nil")
	}
	for _, msg := range cr.Content {
		// fmt.Println(msg.Role)
		// for _, p := range msg.Parts {
		// 	fmt.Println(p)
		// }
		if len(msg.Parts) == 0 {
			return fmt.Errorf("some message has no parts")
		}
	}
	return nil
}

func RestHandler(ctx context.Context, a Agent, e *echo.Echo) {
	if e == nil {
		panic("got nil parameter")
	}

	e.POST("/v1/chat/completions", func(c echo.Context) error {
		slog.Debug("got request")
		if ok := IsJsonContentType(c.Request()); !ok {
			return c.JSON(400, echo.Map{"error": "expecting json body"})
		}

		var input ChatRequest
		// input := map[string]any{}
		if err := c.Bind(&input); err != nil {
			slog.Error("failed binding", "error", err, "type", input)
			return c.JSON(400, echo.Map{"error": "bad json format"})
		}

		if err := input.validate(); err != nil {
			slog.Error("validate error", "error", err)
			return c.JSON(400, echo.Map{"error": "bad json format."})
		}

		output, err := a.Completion(c.Request().Context(), input.Content)

		if err != nil {
			slog.Error("failed completion", "error", err)
			return c.JSON(400, echo.Map{"error": "server unavailable"})
		}

		slog.Debug("request finish")
		return c.JSON(200, ChatResponse{
			Text: output.Text(),
		})
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
