package main

import (
	"flag"
	"os"
	"time"
)

/*
	var hieracrchy:
	1. default (if any)
	2. env
	3. flag
*/

const (
	LLM_ENDPOINT_DEFAULT = "http://localhost:11823"
)

type Config struct {
	Bot BotConfig
	LLM LLMConfig
}

type BotConfig struct {
	IsProd  bool
	Key     string
	Timeout time.Duration
}

type LLMConfig struct {
	Endpoint string
	Key      string
}

func DefaultConfig() Config {
	var conf Config
	conf.Bot.Timeout = 10 * time.Second

	_ = ReadFromEnv(&conf)
	_ = ReadFromFlags(&conf)
	return conf
}

func ReadFromFlags(conf *Config) error {
	prod := flag.Bool("prod", false, "deploy tags")
	key := flag.String("key", "", "telegram bot api key")
	backend := flag.String("backend", "http://localhost:11823", "llm backend endpoint")
	backend_key := flag.String("backend-key", "", "llm backend key")
	flag.Parse()

	if *backend != "" {
		conf.LLM.Endpoint = *backend
	}

	if *backend_key != "" {
		conf.LLM.Endpoint = *backend_key
	}

	if *key != "" {
		conf.Bot.Key = *key
	}
	conf.Bot.IsProd = *prod
	return nil
}

func ReadFromEnv(conf *Config) error {
	tkn := boTokenVar()
	conf.Bot.Key = tkn
	return nil
}

func boTokenVar() string {
	key := os.Getenv("TG_BOT_API_KEY")
	return key
}
