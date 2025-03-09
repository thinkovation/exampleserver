package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// WebhookPlugin forwards log entries to a webhook URL
type WebhookPlugin struct {
	URL    string    `json:"url"`
	APIKey string    `json:"api_key"`
	Filter LogFilter `json:"filter"`
	client *http.Client
}

func NewWebhookPlugin(url, apiKey string, filter LogFilter) *WebhookPlugin {
	return &WebhookPlugin{
		URL:    url,
		APIKey: apiKey,
		Filter: filter,
		client: &http.Client{},
	}
}

func (w *WebhookPlugin) Initialize() error {
	// Validate URL
	if w.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	return nil
}

func (w *WebhookPlugin) Close() error {
	w.client.CloseIdleConnections()
	return nil
}

func (w *WebhookPlugin) ShouldHandle(entry LogEntry) bool {
	// Check levels
	fmt.Println("Checking levels", entry.Level, entry.Message)
	if len(w.Filter.Levels) > 0 {
		levelMatch := false
		for _, level := range w.Filter.Levels {
			if strings.EqualFold(entry.Level, level) {
				fmt.Println("WebhookPlugin ShouldHandle: Level match", level)
				levelMatch = true
				break
			}
		}
		if !levelMatch {
			return false
		}
	}

	// Check sources
	if len(w.Filter.Sources) > 0 {
		sourceMatch := false
		for _, source := range w.Filter.Sources {
			if strings.Contains(entry.Source, source) {
				sourceMatch = true
				break
			}
		}
		if !sourceMatch {
			return false
		}
	}

	// Check contains
	if len(w.Filter.Contains) > 0 {
		for _, substr := range w.Filter.Contains {
			if !strings.Contains(entry.Message, substr) {
				return false
			}
		}
	}

	// Check time range
	if w.Filter.StartTime != nil && entry.Timestamp.Before(*w.Filter.StartTime) {
		return false
	}
	if w.Filter.EndTime != nil && entry.Timestamp.After(*w.Filter.EndTime) {
		return false
	}

	// Check field matches
	if len(w.Filter.FieldMatch) > 0 {
		for key, value := range w.Filter.FieldMatch {
			if fieldValue, ok := entry.Fields[key]; !ok || fieldValue != value {
				return false
			}
		}
	}

	return true
}

func (w *WebhookPlugin) Handle(entry LogEntry) error {
	fmt.Println("Handling webhook", entry.Level, entry.Message)
	// Convert entry to JSON
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", w.URL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Failed to create request", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", w.APIKey)

	// Send request
	resp, err := w.client.Do(req)
	if err != nil {
		fmt.Println("Failed to send webhook", err)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		fmt.Println("Webhook request failed with status", resp.StatusCode)
		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	return nil
}
