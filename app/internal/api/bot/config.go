package bot

import (
	"errors"
	"time"
)

type Config struct {
	Token   string
	Timeout time.Duration
}

var (
	ErrTokenRequired = errors.New("Bot token required")
)

func (c *Config) Validate() error {
	if c.Token == "" {
		return ErrTokenRequired
	}

	if c.Timeout <= 0 {
		c.Timeout = 5 * time.Second
	}
	return nil
}
