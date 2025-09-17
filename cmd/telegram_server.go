package cmd

import (
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/odit-bit/jagatai/api"
	telebot "github.com/odit-bit/jagatai/tgbot"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v4"
)

func init() {
	TeleCMD.Flags().Bool("prod", false, "deployment tags")
}

var TeleCMD = cobra.Command{
	Use: "bot",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		isProd, _ := cmd.Flags().GetBool("prod")
		botConfig := telebot.BotConfig{
			IsProd: isProd,
			Key:    telebot.GetBotTokenEnv(),
		}
		if botConfig.IsProd {
			slog.Info("Deployment", "is_production", botConfig.IsProd)
			slog.SetLogLoggerLevel(slog.LevelError)
		} else {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		//bot
		setting := tele.Settings{
			Token:  botConfig.Key,
			Poller: &tele.LongPoller{Timeout: botConfig.Timeout},
		}
		bot, err := tele.NewBot(setting)
		if err != nil {
			log.Fatal(err)
		}

		// llm backend
		llmConfig := telebot.LLMConfig{}
		ai := api.NewClient(llmConfig.Endpoint, llmConfig.Key)

		// cache
		cache := telebot.NewCache()

		telebot.Handle(ctx, bot, ai, cache)

		srvErr := make(chan error, 1)
		go func() {
			bot.Start()
			_, err := bot.Close()
			srvErr <- err
		}()

		select {
		case err = <-srvErr:
			return err
		case <-ctx.Done():
			stop()
		}

		bot.Stop()

		return nil
	},
}
