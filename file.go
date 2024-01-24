package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
)

type FileStatus string

const (
	FilestatusUploaded  FileStatus = "uploaded"
	FilestatusProcessed FileStatus = "processed"
	FilestatusError     FileStatus = "error"
)

type FilePurpose string

const (
	FilePurposeFinetune         FilePurpose = "fine-tune"
	FilePurposeFinetuneResults  FilePurpose = "fine-tune-results"
	FilePurposeAssistants       FilePurpose = "assistants"
	FilePurposeAssistantsOutput FilePurpose = "assistants_output"
)

type File struct {
	ID            string      `json:"id"`
	Object        string      `json:"object"`
	Bytes         int         `json:"bytes"`
	CreatedAt     int         `json:"created_at"`
	Filename      string      `json:"filename"`
	Purpose       FilePurpose `json:"purpose"`
	Status        FileStatus  `json:"status"`         // Deprecated
	StatusDetails string      `json:"status_details"` // Deprecated
}

type FileList struct {
	Data   []File `json:"data"`
	Object string `json:"object"`
}

type DeleteFileResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// UploadFile implements https://platform.openai.com/docs/api-reference/files/create.
func (c *Client) UploadFile(ctx context.Context, file io.Reader, purpose FilePurpose) (*File, error) {
	endpoint := "/files"
	httpReq, err := http.NewRequest("POST", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	err = writer.WriteField("purpose", string(purpose))
	if err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("file", "mydata.jsonl")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	httpReq.Body = io.NopCloser(&requestBody)

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var responseFile File
	if err := json.NewDecoder(res.Body).Decode(&responseFile); err != nil {
		return nil, err
	}

	return &responseFile, nil
}

// ListFiles implements https://platform.openai.com/docs/api-reference/files/list.
func (c *Client) ListFiles(ctx context.Context) (*FileList, error) {
	endpoint := "/files"
	httpReq, err := http.NewRequest("GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var fileList FileList
	if err := json.NewDecoder(res.Body).Decode(&fileList); err != nil {
		return nil, err
	}

	return &fileList, nil
}

// RetrieveFile implements https://platform.openai.com/docs/api-reference/files/retrieve.
func (c *Client) RetrieveFile(ctx context.Context, fileID string) (*File, error) {
	endpoint := "/files/" + fileID
	httpReq, err := http.NewRequest("GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var file File
	if err := json.NewDecoder(res.Body).Decode(&file); err != nil {
		return nil, err
	}

	return &file, nil
}

// DeleteFile implements https://platform.openai.com/docs/api-reference/files/delete.
func (c *Client) DeleteFile(ctx context.Context, fileID string) (*DeleteFileResponse, error) {
	endpoint := "/files/" + fileID
	httpReq, err := http.NewRequest("DELETE", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var deletedFile DeleteFileResponse
	if err := json.NewDecoder(res.Body).Decode(&deletedFile); err != nil {
		return nil, err
	}

	return &deletedFile, nil
}

// RetrieveFileContent implements https://platform.openai.com/docs/api-reference/files/retrieve-contents.
func (c *Client) RetrieveFileContent(ctx context.Context, fileID string) (string, error) {
	endpoint := "/files/" + fileID + "/content"

	httpReq, err := http.NewRequest("GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return "", err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := c.sendRequest(ctx, httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var fileContent string
	if err := json.NewDecoder(res.Body).Decode(&fileContent); err != nil {
		return "", err
	}

	return fileContent, nil
}
