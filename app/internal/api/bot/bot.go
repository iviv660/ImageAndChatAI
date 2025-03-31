package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/app/internal/service/replicate"
	"time"

	tb "gopkg.in/telebot.v4"
)

type ServiceAI interface {
	ChatCompletion(context.Context, int64, string) (string, error)
	NewConversation(context.Context, int64)
	GenerateImagePrompt(ctx context.Context, prompt string) (string, error)
}

type replicateService interface {
	GenerateImage(ctx context.Context, reqGen *replicate.Request) (replicate.Response, error)
}

type Wrapper struct {
	bot         *tb.Bot
	config      *Config
	openai      ServiceAI
	replication replicateService
}

var Us = make(map[int64]int)

func NewWrapper(cfg *Config, openai ServiceAI, replication replicateService) (*Wrapper, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if openai == nil {
		return nil, errors.New("openai service cannot be nil")
	}

	settings := tb.Settings{
		Token:  cfg.Token,
		Poller: &tb.LongPoller{Timeout: cfg.Timeout},
		OnError: func(err error, c tb.Context) {
			log.Printf("Telebot error: %v", err)
			if c != nil {
				c.Send("Произошла ошибка, попробуйте позже")
			}
		},
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	w := &Wrapper{
		bot:         bot,
		config:      cfg,
		openai:      openai,
		replication: replication,
	}

	w.prepareHandlers()

	return w, nil
}

func (w *Wrapper) Start() {
	if w.bot == nil {
		log.Fatal("Bot is not initialized")
	}
	log.Println("Starting bot...")
	w.bot.Start()
}

func (w *Wrapper) prepareHandlers() {
	w.bot.Handle("/start", func(c tb.Context) error {
		menu := &tb.ReplyMarkup{
			ResizeKeyboard: true,
			ReplyKeyboard: [][]tb.ReplyButton{
				{{Text: "🆕 Новый Чат"}, {Text: "🖼 Генерация изображения"}},
			},
		}
		Us[c.Sender().ID] = 1
		return c.Send("Выберите действие:", menu)
	})

	w.bot.Handle("🆕 Новый Чат", func(c tb.Context) error {
		if Us[c.Sender().ID] == 1 {
			ctx := context.TODO()

			w.openai.NewConversation(ctx, c.Sender().ID)

			next := &tb.ReplyMarkup{
				ResizeKeyboard: true,
				ReplyKeyboard: [][]tb.ReplyButton{
					{{Text: "Назад"}},
				},
			}

			Us[c.Sender().ID] = 2
			return c.Send("Новый диалог начат. Что вы хотите обсудить?", next)
		}
		return nil
	})

	w.bot.Handle("🖼 Генерация изображения", func(c tb.Context) error {
		if Us[c.Sender().ID] == 1 {
			ctx := context.TODO()
			w.openai.NewConversation(ctx, c.Sender().ID)

			next := &tb.ReplyMarkup{
				ResizeKeyboard: true,
				ReplyKeyboard: [][]tb.ReplyButton{
					{{Text: "Назад"}},
				},
			}

			return c.Send("Функция генерации изображений включена", next)
		}
		return nil
	})

	w.bot.Handle("Назад", func(c tb.Context) error {
		menu := &tb.ReplyMarkup{
			ResizeKeyboard: true,
			ReplyKeyboard: [][]tb.ReplyButton{
				{{Text: "🆕 Новый Чат"}, {Text: "🖼 Генерация изображения"}},
			},
		}
		Us[c.Sender().ID] = 1
		return c.Send("Выберите действие:", menu)
	})

	w.bot.Handle(tb.OnText, func(c tb.Context) error {
		if Us[c.Sender().ID] == 2 {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := c.Notify(tb.Typing)
			if err != nil {
				log.Printf("Failed to show typing indicator: %v", err)
			}

			msg, err := w.openai.ChatCompletion(ctx, c.Sender().ID, c.Text())
			if err != nil {
				log.Printf("ChatCompletion error: %v", err)
				return c.Send("Извините, произошла ошибка при обработке вашего запроса. Попробуйте позже.")
			}

			maxLength := 4000
			if len(msg) > maxLength {
				for i := 0; i < len(msg); i += maxLength {
					end := i + maxLength
					if end > len(msg) {
						end = len(msg)
					}
					if err := c.Send(msg[i:end]); err != nil {
						return err
					}
				}
				return nil
			}

			return c.Send(msg)
		}
		if Us[c.Sender().ID] == 3 {
			ctx := context.TODO()
			openaiPrompt, err := w.openai.GenerateImagePrompt(ctx, c.Text())
			res, err := w.replication.GenerateImage(ctx, &replicate.Request{Input: &replicate.Input{
				Prompt: openaiPrompt,
				Ratio:  "16:9",
			}})
			if err != nil {
				log.Printf("ChatCompletion error: %v", err)
				return c.Send("Извините, произошла ошибка при обработке вашего запроса. Попробуйте позже.")
			}
			photo := &tb.Photo{
				File: tb.FromURL(res.Output),
			}
			return c.Send(photo)
		}
		return c.Send("Вы делаете что-то не так!")
	})
}
