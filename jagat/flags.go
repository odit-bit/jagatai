package jagat

import "github.com/spf13/pflag"

const (
	FLAG_PROVIDER_KEY      = "p_key"
	FLAG_PROVIDER_ENDPOINT = "p_addr"
	FLAG_PROVIDER_NAME     = "p_name"
	FLAG_PROVIDER_MODEL    = "p_model"

	FLAG_SERVER_ADDRESS     = "addr"
	FLAG_SERVER_DEBUG       = "debug"
	FLAG_SERVER_CONFIG_FILE = "config"

	FLAG_OBSERVE_ENABLE         = "observe"
	FLAG_OBSERVE_TRACEENDPOINT  = "traceendpoint"
	FLAG_OBSERVE_METER_ENDPOINT = "metricendpoint"
)

// Defined set of flags for jagat configuration use.
var FlagSet = pflag.NewFlagSet("Jagat_Flags", pflag.PanicOnError)

var flagToConfigKeyMap = map[string]string{
	FLAG_PROVIDER_KEY:      "provider.apikey",
	FLAG_PROVIDER_ENDPOINT: "provider.endpoint",
	FLAG_PROVIDER_NAME:     "provider.name",
	FLAG_PROVIDER_MODEL:    "provider.model",

	FLAG_SERVER_ADDRESS: "server.address",
	FLAG_SERVER_DEBUG:   "server.debug",
	// FLAG_SERVER_CONFIG_FILE: "config",

	FLAG_OBSERVE_ENABLE:        "observe.enable",
	FLAG_OBSERVE_TRACEENDPOINT: "",
}

func init() {
	defineFlags()
}

func defineFlags() {
	// server
	FlagSet.String(FLAG_SERVER_ADDRESS, "", "server address")
	FlagSet.Bool(FLAG_SERVER_DEBUG, false, "debug log")
	FlagSet.String(FLAG_SERVER_CONFIG_FILE, "", "path to config file")

	// provider
	FlagSet.String(FLAG_PROVIDER_KEY, "", "provider's api key")
	FlagSet.String(FLAG_PROVIDER_NAME, "", "provider's name")
	FlagSet.String(FLAG_PROVIDER_MODEL, "", "provider's model name")

	//observe
	FlagSet.Bool(FLAG_OBSERVE_ENABLE, false, "enable observability default false")
}
