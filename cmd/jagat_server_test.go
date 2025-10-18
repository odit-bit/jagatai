package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Tests for LoadAndValidate function
// =============================================================================

// Helper function to create a new FlagSet for isolated tests
func newTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String(FLAG_SERVER_ADDRESS, "", "server address")
	flags.Bool(FLAG_SERVER_DEBUG, false, "debug log")
	flags.String(FLAG_SERVER_CONFIG_FILE, "", "path to config file")
	flags.String(FLAG_PROVIDER_KEY, "", "provider's api key")
	flags.String(FLAG_PROVIDER_NAME, "", "provider's name")
	flags.String(FLAG_PROVIDER_MODEL, "", "base model agent use")
	return flags
}

func TestLoadAndValidate_Old(t *testing.T) {
	// --- Test Case 1: Load from default config file ---
	t.Run("loads from default config", func(t *testing.T) {
		flags := newTestFlagSet()
		cfg, err := LoadAndValidate(flags) // Assumes LoadAndValidate can find the embedded default

		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:11823", cfg.Server.Address)
		assert.Equal(t, "ollama", cfg.Provider.Name)
		assert.Equal(t, "qwen3:1.7b", cfg.Provider.Model)
		assert.False(t, cfg.Server.Debug)
	})

	// --- Test Case 2: Flag overrides config file ---
	t.Run("flag overrides config file", func(t *testing.T) {
		flags := newTestFlagSet()
		flags.Parse([]string{"--server.address", "localhost:9999", "--server.debug=true"})
		cfg, err := LoadAndValidate(flags)

		require.NoError(t, err)
		assert.Equal(t, "localhost:9999", cfg.Server.Address) // Overridden
		assert.True(t, cfg.Server.Debug)                      // Overridden
		assert.Equal(t, "ollama", cfg.Provider.Name)          // From default config
	})

	// --- Test Case 3: Environment variable overrides config file ---
	t.Run("env var overrides config file", func(t *testing.T) {
		// Set environment variables
		t.Setenv("JAGATAI_PROVIDER_NAME", "openai")
		t.Setenv("JAGATAI_SERVER_DEBUG", "true")
		t.Setenv("JAGATAI_PROVIDER_APIKEY", "apikey_value")

		flags := newTestFlagSet()
		cfg, err := LoadAndValidate(flags)

		require.NoError(t, err)
		assert.Equal(t, "openai", cfg.Provider.Name)         // Overridden by env
		assert.Equal(t, "apikey_value", cfg.Provider.ApiKey) // Overridden by env
		assert.True(t, cfg.Server.Debug)                     // Overridden by env
	})

	// --- Test Case 4: Flag overrides both env var and config file ---
	t.Run("flag overrides env var and config", func(t *testing.T) {
		t.Setenv("JAGATAI_PROVIDER_MODEL", "gpt-3.5-turbo")
		t.Setenv("JAGATAI_SERVER_ADDRESS", "0.0.0.0:8080")

		flags := newTestFlagSet()
		flags.Parse([]string{"--provider.model", "claude-3"})

		cfg, err := LoadAndValidate(flags)

		require.NoError(t, err)
		assert.Equal(t, "claude-3", cfg.Provider.Model)     // Overridden by flag
		assert.Equal(t, "0.0.0.0:8080", cfg.Server.Address) // From env var (flag not set)
	})

	// 	// --- Test Case 5: Validation error for missing required field ---
	// 	t.Run("validation fails for missing provider name", func(t *testing.T) {
	// 		// Create a temporary, minimal config file
	// 		content := []byte(`
	// server:
	//   address: "127.0.0.1:1234"
	// provider:
	//   model: "a-model"
	// # name is missing
	// `)
	// 		tmpfile, err := os.CreateTemp("", "config-*.yaml")
	// 		require.NoError(t, err)
	// 		defer os.Remove(tmpfile.Name())
	// 		_, err = tmpfile.Write(content)
	// 		require.NoError(t, err)
	// 		tmpfile.Close()

	// 		flags := newTestFlagSet()
	// 		flags.Parse([]string{"--config", tmpfile.Name()})

	// 		_, err = LoadAndValidate(flags)
	// 		require.Error(t, err)
	// 		assert.Contains(t, err.Error(), "provider name is required")
	// 	})
}
