package kong

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KongAdminClient handles communication with Kong Admin API
type KongAdminClient struct {
	baseURL    string
	httpClient *http.Client
}

// ConsumerRequest represents the request to create a Kong consumer
type ConsumerRequest struct {
	Username  string            `json:"username"`
	CustomID  string            `json:"custom_id,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ConsumerResponse represents the response from Kong when creating a consumer
type ConsumerResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CustomID  string    `json:"custom_id"`
	CreatedAt int64     `json:"created_at"`
	Tags      []string  `json:"tags,omitempty"`
}

// KongError represents an error response from Kong
type KongError struct {
	Message string `json:"message"`
	Name    string `json:"name"`
	Code    int    `json:"code"`
}

func (e *KongError) Error() string {
	return fmt.Sprintf("Kong API error [%d]: %s - %s", e.Code, e.Name, e.Message)
}

// NewKongAdminClient creates a new Kong Admin API client
func NewKongAdminClient(baseURL string, timeout int) *KongAdminClient {
	return &KongAdminClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// CreateConsumer creates a new consumer in Kong
// The consumer username should be the user's UUID
func (k *KongAdminClient) CreateConsumer(ctx context.Context, req *ConsumerRequest) (*ConsumerResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consumer request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", k.baseURL+"/consumers", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Kong: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusCreated {
		var kongErr KongError
		if err := json.Unmarshal(body, &kongErr); err != nil {
			return nil, fmt.Errorf("failed to create consumer in Kong (status %d): %s", resp.StatusCode, string(body))
		}
		kongErr.Code = resp.StatusCode
		return nil, &kongErr
	}

	var consumer ConsumerResponse
	if err := json.Unmarshal(body, &consumer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consumer response: %w", err)
	}

	return &consumer, nil
}

// DeleteConsumer deletes a consumer from Kong by username or ID
func (k *KongAdminClient) DeleteConsumer(ctx context.Context, usernameOrID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", k.baseURL+"/consumers/"+usernameOrID, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to Kong: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete consumer in Kong (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetConsumer retrieves a consumer from Kong by username or ID
func (k *KongAdminClient) GetConsumer(ctx context.Context, usernameOrID string) (*ConsumerResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", k.baseURL+"/consumers/"+usernameOrID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Kong: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var kongErr KongError
		if err := json.Unmarshal(body, &kongErr); err != nil {
			return nil, fmt.Errorf("failed to get consumer from Kong (status %d): %s", resp.StatusCode, string(body))
		}
		kongErr.Code = resp.StatusCode
		return nil, &kongErr
	}

	var consumer ConsumerResponse
	if err := json.Unmarshal(body, &consumer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consumer response: %w", err)
	}

	return &consumer, nil
}
