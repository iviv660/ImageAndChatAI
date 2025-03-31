package replicate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

var ErrBadRequest = errors.New("bad request")

type Service struct {
	token, url string
}

func NewService(cfg *Config) (*Service, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	url := "https://api.replicate.com/v1/models/black-forest-labs/flux-1.1-pro-ultra/predictions"

	return &Service{token: cfg.Token, url: url}, nil
}

func (s *Service) GenerateImage(ctx context.Context, reqGen *Request) (Response, error) {
	data, err := json.Marshal(reqGen)
	if err != nil {
		return Response{}, err
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewBuffer(data))
	if err != nil {
		return Response{}, err
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}

	defer func() {
		errBodyClose := resp.Body.Close()
		if errBodyClose != nil {
			slog.Error("failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		responseBody := ""
		if body != nil {
			responseBody = string(body)
		}
		slog.With(
			slog.String("status", resp.Status),
			slog.Int("status_code", resp.StatusCode),
			slog.String("body", responseBody),
		).Error("Image generation request failed")

		// Returning error response
		return Response{}, ErrBadRequest
	}

	// Парсим ответ
	var res Response
	if err = json.Unmarshal(body, &res); err != nil {
		return Response{}, err
	}

	return res, nil
}
