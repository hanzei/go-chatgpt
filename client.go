package chatgpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	apiURL = "https://api.openai.com/v1"
)

type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Config
	config *Config
}

type Config struct {
	// Base URL for API requests.
	BaseURL string

	// API Key (Required)
	APIKey string

	// Organization ID (Optional)
	OrganizationID string
}

var (
	// ErrAPIKeyRequired is returned when the API Key is not provided
	ErrAPIKeyRequired = errors.New("API Key is required")

	// ErrInvalidModel is returned when the model is invalid
	ErrInvalidModel = errors.New("invalid model")

	// ErrNoMessages is returned when no messages are provided
	ErrNoMessages = errors.New("no messages provided")

	// ErrInvalidRole is returned when the role is invalid
	ErrInvalidRole = errors.New("invalid role. Only `user`, `system` and `assistant` are supported")

	// ErrInvalidTemperature is returned when the temperature is invalid
	ErrInvalidTemperature = errors.New("invalid temperature. 0<= temp <= 2")

	// ErrInvalidPresencePenalty
	ErrInvalidPresencePenalty = errors.New("invalid presence penalty. -2<= presence penalty <= 2")

	// ErrInvalidFrequencyPenalty
	ErrInvalidFrequencyPenalty = errors.New("invalid frequency penalty. -2<= frequency penalty <= 2")
)

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   any    `json:"param"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func NewClient(apikey string) (*Client, error) {
	if apikey == "" {
		return nil, ErrAPIKeyRequired
	}

	return &Client{
		client: &http.Client{},
		config: &Config{
			BaseURL: apiURL,
			APIKey:  apikey,
		},
	}, nil
}

func NewClientWithConfig(config *Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, ErrAPIKeyRequired
	}

	return &Client{
		client: &http.Client{},
		config: config,
	}, nil
}

func (c *Client) sendRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	if c.config.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrganizationID)
	}

	// Default to Content-Type json
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		// Parse body
		var apiError apiError
		if err := json.NewDecoder(res.Body).Decode(&apiError); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("api request failed: status Code: %d, URL: %s, Error type: %s, Message: %s",
			res.StatusCode,
			res.Request.URL,
			apiError.Error.Type,
			apiError.Error.Message,
		)
	}

	return res, nil
}
