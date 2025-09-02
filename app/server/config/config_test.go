package config

import (
	"testing"
)

func Test_config(t *testing.T) {
	// v := []tooldef.Config{}
	cfg := Config{}
	err := UnmarshalConfigFile("", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	tool1 := cfg.Tools[0]
	if tool1.Name != "openmeteo" {
		t.Fatal("different name")
	}
	if tool1.Endpoint != "https://api.open-meteo.com" {
		t.Fatal("different endpoint")
	}

}
