package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/odit-bit/jagatai/client"
	tele "gopkg.in/telebot.v4"
)

var sysPromp = "You are usefull assistant, ignore tools or functions for factual questions. Only use tools or function for relevant question. Don't mention if you use tool or not. /no_think"
var sysMsg = client.Message{
	Role:    "system",
	Content: sysPromp,
}

func HandleBot(ctx context.Context, bot *tele.Bot, llmClient *client.Client, cache *ChatCache) {
	//bot
	bot.Handle("/start", func(ctx tele.Context) error {
		slog.Info("GOT Start")

		return ctx.Send("hi..")
	})

	bot.Handle(tele.OnLocation, func(ctx tele.Context) error {
		slog.Info("GOT location")
		if ctx.Message().Location.Lat != 0 {
			var input string
			loc := ctx.Message().Location
			input = fmt.Sprintf("Lat:%f, Long:%f", loc.Lat, loc.Lng)

			msg, err := Completion(ctx.Chat().ID, cache, llmClient, input)
			if err != nil {
			}
			if err != nil {
				slog.Error("location error", "error", err.Error())
				return ctx.Send("service unavailable")
			}
			return ctx.Send(msg)
		}
		return ctx.Send("empty location")
	})

	//text Handler
	bot.Handle(tele.OnText, func(ctx tele.Context) error {
		slog.Info("GOT TEXT")

		/*store chat ID*/

		msg, err := Completion(ctx.Chat().ID, cache, llmClient, ctx.Message().Text)
		if err != nil {
		}
		if err != nil {
			slog.Error("text error", "error", err.Error())
			return ctx.Send("service unavailable")
		}

		if err := ctx.Send(msg); err != nil {
			return err
		}
		return nil

		// res := client.ChatResponse{
		// 	Message: client.Message{
		// 		Role:    "assistant",
		// 		Content: fmt.Sprintf("mock message %s", time.Now().String()),
		// 	},
		// }
		// return ctx.Send(res.Message.Content)

	})

	bot.Handle("/count", func(ctx tele.Context) error {
		slog.Info("GOT count")

		n := cache.CountMessages(ctx.Chat().ID)
		return ctx.Send(fmt.Sprintf("%d", n))
	})

	bot.Handle("/clear", func(ctx tele.Context) error {
		slog.Info("GOT clear")

		_ = cache.Clear(ctx.Chat().ID)
		return ctx.Send("context clear")
	})
}

func Completion(id int64, cache *ChatCache, ai *client.Client, content string) (string, error) {
	/*store chat ID*/

	sc := cache.Get(id)
	sc.Add(client.Message{Role: "user", Content: content})

	resp, err := ai.Chat(client.ChatRequest{
		Messages: append(append([]client.Message{}, sysMsg), sc.Messages()...),
	})
	if err != nil {
		return "", err
	}

	resp.Message.Content = ParseThink(resp.Message.Content)

	sc.Add(resp.Message)
	sc.Save()
	return resp.Message.Content, nil
}

func ParseThink(msg string) string {
	close := "</think>"
	idx := strings.Index(msg, close)
	if idx != -1 {
		return strings.TrimSpace(msg[idx+len(close):])
	}
	return msg
}
