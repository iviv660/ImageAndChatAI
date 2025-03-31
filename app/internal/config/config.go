package config

import (
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Bot       BotConfig       `yaml:"bot"`
	OpenAI    OpenAIConfig    `yaml:"openai"`
	Replicate ReplicateConfig `yaml:"replicate"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Tracing   TracingConfig   `yaml:"tracing"`
}

type AppConfig struct {
	Id   string `yaml:"id" env:"APP_ID"`
	Name string `yaml:"name" env:"APP_NAME"`
}

type BotConfig struct {
	Token   string        `yaml:"token" env:"BOT_TOKEN"`
	Timeout time.Duration `yaml:"timeout" env:"BOT_TIMEOUT"`
}

type OpenAIConfig struct {
	Enabled bool   `yaml:"enabled" env:"OPENAI_ENABLED"`
	ApiKey  string `yaml:"api_key" env:"OPENAI_API_KEY"`
}

type ReplicateConfig struct {
	Enabled bool   `yaml:"enabled" env:"REPLICATE_ENABLED"`
	Token   string `yaml:"token" env:"REPLICATE_TOKEN"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Host    string `yaml:"host" env:"METRICS_HOST"`
	Port    int    `yaml:"port" env:"METRICS_PORT"`
}

type TracingConfig struct {
	Enabled bool   `yaml:"enabled" env:"TRACING_ENABLED"`
	Host    string `yaml:"host" env:"TRACING_HOST"`
	Port    int    `yaml:"port" env:"TRACING_PORT"`
}

const (
	flagConfigPathName = "config"
	envConfigPathName  = "CONFIG_PATH"
)

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		var configPath string
		flag.StringVar(&configPath, flagConfigPathName, "", "path to config file")
		flag.Parse()

		// Используем переменную окружения, если она задана
		if path, ok := os.LookupEnv(envConfigPathName); ok {
			configPath = path
		}

		// Если путь не задан, устанавливаем дефолтный
		if configPath == "" {
			configPath = "C:/dev/projekts/UI images/configs/config.yml"
		}

		// Логируем путь к конфигурационному файлу
		slog.Info("Using config path", slog.String("config_path", configPath))

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			slog.Error("Configuration file does not exist", slog.String("config_path", configPath))
			os.Exit(1)
		}

		instance = &Config{}

		// Загружаем конфигурацию
		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			slog.Error("Failed to read configuration", slog.String("error", err.Error()))
			os.Exit(1)
		}

		// Проверяем, не nil ли instance
		if instance == nil {
			slog.Error("Configuration instance is nil")
			os.Exit(1)
		}

		// Проверяем, инициализированы ли вложенные структуры
		if (Config{}) == *instance {
			slog.Error("Loaded configuration is empty")
			os.Exit(1)
		}

		// Проверяем обязательные поля
		if instance.App.Id == "" || instance.App.Name == "" {
			slog.Error("Application configuration is invalid",
				slog.String("app_id", instance.App.Id),
				slog.String("app_name", instance.App.Name),
			)
			os.Exit(1)
		}

		slog.Info("Configuration loaded successfully")
	})

	return instance
}
