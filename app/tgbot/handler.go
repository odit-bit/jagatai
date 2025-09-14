package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/client"
	tele "gopkg.in/telebot.v4"
)

var sysPromp = "You are usefull assistant, ignore tools or functions for factual questions. Only use tools or function for relevant question. Don't mention if you use tool or not."
var sysMsg = client.Message{
	Role: "system",
	Text: sysPromp,
}

func HandleBot(ctx context.Context, bot *tele.Bot, llmClient *client.Client, cache *ChatCache) {
	//bot
	bot.Handle("/start", func(ctx tele.Context) error {
		slog.Info("GOT Start")
		return ctx.Send("hi..")
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

	h := Handler{
		ctx:   ctx,
		ai:    llmClient,
		cache: cache,
	}

	bot.Handle(tele.OnText, h.HandleText)
	bot.Handle(tele.OnPhoto, h.HandlePhoto)
	bot.Handle(tele.OnLocation, h.HandleLoc)
	bot.Handle(tele.OnDocument, h.HandleDoc)
}

type Handler struct {
	ctx   context.Context
	ai    *client.Client
	cache *ChatCache
}

func (h *Handler) HandleDoc(ctx tele.Context) error {
	doc := ctx.Message().Document
	if doc.MIME != "application/pdf" {
		return ctx.Send("file only support pdf")
	}
	f, err := ctx.Bot().File(&doc.File)
	if err != nil {
		slog.Error("failed to get doc from telegram", "error", err)
		return ctx.Send("server error")
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		slog.Error("failed read doc", "error", err)
		return ctx.Send("server error")
	}

	resp, err := h.do(
		h.ctx,
		ctx.Message().Chat.ID,
		&client.Message{
			Role: "user",
			Data: &agent.Blob{
				Bytes: b,
				Mime:  doc.MIME,
			},
		},
	)
	if err != nil {
		slog.Error("failed generate content", "error", err)
		return ctx.Send("server errror")
	}
	return ctx.Send(resp.Message.Text)
}

func (h *Handler) HandleText(ctx tele.Context) error {
	slog.Info("GOT TEXT")

	/*store chat ID*/

	res, err := h.do(h.ctx, ctx.Chat().ID, &client.Message{
		Role: "user",
		Text: ctx.Text(),
	})
	if err != nil {
		return ctx.Send("service unavailable")
	}
	return ctx.Send(res.Message.Text)
}

func (h *Handler) HandleLoc(ctx tele.Context) error {
	slog.Info("GOT Location")

	res, err := h.do(h.ctx, ctx.Chat().ID, &client.Message{
		Role: "user",
		Text: fmt.Sprintf(
			"Lat:%f, Long:%f",
			ctx.Message().Location.Lat,
			ctx.Message().Location.Lng,
		),
	})
	if err != nil {
		return ctx.Send("service unavailable")
	}
	return ctx.Send(res.Message.Text)
}

func (h *Handler) HandlePhoto(ctx tele.Context) error {
	photo := ctx.Message().Photo

	if photo.File.InCloud() {

		// b, _ := json.MarshalIndent(photo, "", " ")
		// fmt.Println(string(b))

		rc, err := ctx.Bot().File(&photo.File)
		if err != nil {
			slog.Error(fmt.Errorf("server failed to fetch file: %v", err).Error())
			return ctx.Send("server error ")
		}
		defer rc.Close()

		b, err := io.ReadAll(rc)
		if err != nil {
			return ctx.Send("error")
		}
		mime := http.DetectContentType(b)

		res, err := h.do(h.ctx, ctx.Chat().ID, &client.Message{
			Role: "user",
			Text: photo.InputMedia().Caption,
			Data: &agent.Blob{
				Bytes: b,
				Mime:  mime,
			},
		})
		if err != nil {
			slog.Error(err.Error())
			return ctx.Send("error")
		}
		return ctx.Send(res.Message.Text)
	}

	return ctx.Send("picture not from telegram server")
}

func (h *Handler) do(ctx context.Context, id int64, query *client.Message) (*client.ChatResponse, error) {
	sc := h.cache.Get(id)
	if sc.Len() == 0 {
		sc.Add(sysMsg)
	}
	sc.Add(*query)
	resp, err := h.ai.Chat(
		ctx,
		client.ChatRequest{
			Messages: sc.Messages(),
		},
	)
	if err != nil {
		return nil, err
	}

	resp.Message.Text = ParseThink(resp.Message.Text)

	sc.Add(resp.Message)
	sc.Save()
	return resp, nil
}

func ParseThink(msg string) string {
	close := "</think>"
	idx := strings.Index(msg, close)
	if idx != -1 {
		return strings.TrimSpace(msg[idx+len(close):])
	}
	return msg
}
