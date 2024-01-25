package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type FineTuningJobStatus string

const (
	FineTuningJobStatusValidatingFile FineTuningJobStatus = "validating_files"
	FineTuningJobStatusQueued         FineTuningJobStatus = "queued"
	FineTuningJobStatusRunning        FineTuningJobStatus = "running"
	FineTuningJobStatusSucceeded      FineTuningJobStatus = "succeeded"
	FineTuningJobStatusFailed         FineTuningJobStatus = "failed"
	FineTuningJobStatusCancelled      FineTuningJobStatus = "cancelled"
)

type FineTuningJob struct {
	ID        string `json:"id"`
	CreatedAt int    `json:"created_at"`
	Error     struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param,omitempty"`
	} `json:"error,omitempty"`
	FineTunedModel  string `json:"fine_tuned_model,omitempty"`
	FinishedAt      int    `json:"finished_at,omitempty"`
	Hyperparameters struct {
		NEpochs int `json:"n_epochs"`
	} `json:"hyperparameters,omitempty"`
	Model          string              `json:"model,omitempty"`
	Object         string              `json:"object"`
	OrganizationID string              `json:"organization_id"`
	ResultFiles    []string            `json:"result_files"`
	Status         FineTuningJobStatus `json:"status"`
	TrainedTokens  int                 `json:"trained_tokens,omitempty"`
	TrainingFile   string              `json:"training_file"`
	ValidationFile string              `json:"validation_file,omitempty"`
}

type FineTuningRequest struct {
	Model           ChatGPTModel `json:"model"`
	TrainingFile    string       `json:"training_file"`
	Hyperparameters struct {
		NEpochs int `json:"n_epochs,omitempty"` // Optional
	} `json:"hyperparameters,omitempty"` // Optional
	Suffix         string `json:"suffix,omitempty"`          // Optional
	ValidationFile string `json:"validation_file,omitempty"` // Optional
}

type FineTuningResponse struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	Model          string `json:"model"`
	CreatedAt      int    `json:"created_at"`
	FineTunedModel any    `json:"fine_tuned_model"`
	OrganizationID string `json:"organization_id"`
	ResultFiles    []any  `json:"result_files"`
	Status         string `json:"status"`
	ValidationFile any    `json:"validation_file"`
	TrainingFile   string `json:"training_file"`
}

// TODO: Use generics to create an abstract List type.
type FineTuningList struct {
	Object  string          `json:"object"`
	Data    []FineTuningJob `json:"data"`
	HasMore bool            `json:"has_more"`
}

type FineTuningEvent struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CreatedAt int    `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Type      string `json:"type"`
}

// TODO: Use generics to create an abstract List type.
type FineTuningEventsList struct {
	Object  string            `json:"object"`
	Data    []FineTuningEvent `json:"data"`
	HasMore bool              `json:"has_more"`
}

// CreateFineTuningRequest implements https://platform.openai.com/docs/api-reference/fine-tuning/create.
func (c *Client) CreateFineTuningRequest(ctx context.Context, req FineTuningRequest) (*FineTuningResponse, error) {
	endpoint := "/fine_tuning/jobs"

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response FineTuningResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListFineTuningJobs implements https://platform.openai.com/docs/api-reference/fine-tuning/list.
func (c *Client) ListFineTuningJobs(ctx context.Context, opts *ListOptions) (*FineTuningList, error) {
	endpoint := "/fine_tuning/jobs" + opts.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var fineTuningList FineTuningList
	if err := json.NewDecoder(res.Body).Decode(&fineTuningList); err != nil {
		return nil, err
	}

	return &fineTuningList, nil
}

// ListFineTuningEvents implements https://platform.openai.com/docs/api-reference/fine-tuning/list-events.
func (c *Client) ListFineTuningEvents(ctx context.Context, fineTuningJobID string, opts *ListOptions) (*FineTuningEventsList, error) {
	endpoint := "/fine_tuning/jobs/" + fineTuningJobID + "/events" + opts.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var fineTuningList FineTuningEventsList
	if err := json.NewDecoder(res.Body).Decode(&fineTuningList); err != nil {
		return nil, err
	}

	return &fineTuningList, nil
}

// RetrieveFineTuningJob implements https://platform.openai.com/docs/api-reference/fine-tuning/retrieve.
func (c *Client) RetrieveFineTuningJob(ctx context.Context, fineTuningJobID string) (*FineTuningJob, error) {
	endpoint := "/fine_tuning/jobs/" + fineTuningJobID

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var job FineTuningJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

// CancelFineTuningJob implements https://platform.openai.com/docs/api-reference/fine-tuning/cancel.
func (c *Client) CancelFineTuningJob(ctx context.Context, fineTuningJobID string) (*FineTuningJob, error) {
	endpoint := "/fine_tuning/jobs/" + fineTuningJobID + "/cancel"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var job FineTuningJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}
