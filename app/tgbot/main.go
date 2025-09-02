package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/odit-bit/jagatai/client"
	tele "gopkg.in/telebot.v4"
)



func main() {

	conf := DefaultConfig()
	if conf.Bot.IsProd {
		slog.Info("Deployment", "is_production", conf.Bot.IsProd)
		slog.SetLogLoggerLevel(slog.LevelError)
	} else {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	//bot
	setting := tele.Settings{
		Token:  conf.Bot.Key,
		Poller: &tele.LongPoller{Timeout: conf.Bot.Timeout},
	}
	bot, err := tele.NewBot(setting)
	if err != nil {
		log.Fatal(err)
	}

	// ai backend
	ai := client.NewClient(conf.LLM.Endpoint, conf.LLM.Key)

	// cache
	cache := NewCache()

	HandleBot(context.TODO(), bot, ai, cache)

	bot.Start()
}
