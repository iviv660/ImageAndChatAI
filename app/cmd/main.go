package main

import (
	"fmt"
	"log"
	"log/slog"
	"main/app/internal/api/bot"
	"main/app/internal/config"
	"main/app/internal/service/openai"
	"main/app/internal/service/replicate"
	"os"
)

func main() {

	cfg := config.GetConfig()
	fmt.Println(cfg)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	logger.Info("Application initializing with configuration",
		slog.Any("app_id", cfg.App.Id),
		slog.Any("app_name", cfg.App.Name),
		slog.Any("bot_token_length", len(cfg.Bot.Token)),
		slog.Any("bot_timeout", cfg.Bot.Timeout),
		slog.Any("openai_ebabled", cfg.OpenAI.Enabled),
		slog.Any("openai_apikey", len(cfg.OpenAI.ApiKey)),
		slog.Any("metrics_enabled", cfg.Metrics.Enabled),
		slog.Any("metrics_host", cfg.Metrics.Host),
		slog.Any("metrics_port", cfg.Metrics.Port),
		slog.Any("tracing_enabled", cfg.Tracing.Enabled),
		slog.Any("tracing_host", cfg.Tracing.Host),
		slog.Any("tracing_port", cfg.Tracing.Port),
	)

	DS, err := openai.NewService(&openai.Config{Token: cfg.OpenAI.ApiKey})
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("GigaChat service initialized successfully")

	rep, err := replicate.NewService(&replicate.Config{Token: cfg.Replicate.Token})
	if err != nil {
		log.Fatal(err)
	}

	botWrapper, err := bot.NewWrapper(&bot.Config{Token: cfg.Bot.Token, Timeout: cfg.Bot.Timeout}, DS, rep)
	if err != nil {
		logger.Error("Error creating bot", "error", err)
		os.Exit(1)
	}

	logger.Info("Bot stating...")
	botWrapper.Start()
	logger.Info("Bot run")
}
