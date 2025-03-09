// @title Logger API
// @description API endpoints for managing and retrieving logs
// @version 1.0

package logger

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DebugSettings represents the request body for setting debug mode
// @Description Settings for enabling/disabling debug logging
type DebugSettings struct {
	// Whether debug logging is enabled
	// @Example true
	Enabled bool `json:"enabled"`
}

// LogRequest represents the request parameters for retrieving logs
// @Description Parameters for filtering and formatting log entries
type LogRequest struct {
	// Start time for log filtering (RFC3339 format)
	// @Example 2024-03-10T15:04:05Z
	FromTime *time.Time `json:"from_time,omitempty"`

	// End time for log filtering (RFC3339 format)
	// @Example 2024-03-10T16:04:05Z
	ToTime *time.Time `json:"to_time,omitempty"`

	// Number of most recent log lines to return
	// @Example 100
	LastLines *int `json:"last_lines,omitempty"`

	// Number of minutes of recent logs to return
	// @Example 30
	LastMinutes *int `json:"last_minutes,omitempty"`

	// Output format (json, jsonpretty, csv, text)
	// @Example json
	Format string `json:"format,omitempty"`
}

// LogResponse represents the response for log retrieval
// @Description Collection of log entries
type LogResponse struct {
	// Array of log lines
	Lines []string `json:"lines"`
}

// HTTPHandler manages HTTP endpoints for log operations
type HTTPHandler struct {
	logger LoggerInterface
}

// NewHTTPHandler creates a new logging handler
func NewHTTPHandler(logger LoggerInterface) *HTTPHandler {
	return &HTTPHandler{
		logger: logger,
	}
}

// SetDebug handles requests to change debug logging state
// @Summary Set debug logging mode
// @Description Enable or disable debug logging
// @Tags logger
// @Accept json
// @Produce json
// @Param settings body DebugSettings true "Debug settings"
// @Success 200 {object} DebugSettings
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 405 {string} string "Method not allowed"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/loggersettings/debug [post]
func (h *HTTPHandler) SetDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings DebugSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.SetDebug(settings.Enabled)
	h.logger.Info("Debug logging set to: %v", settings.Enabled)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(settings)
}

// GetLogs handles requests to retrieve log entries
// @Summary Retrieve log entries
// @Description Get filtered log entries with various output formats
// @Tags logger
// @Accept json
// @Produce json,text/csv,text/plain
// @Param from_time query string false "Start time (RFC3339)" Format(date-time)
// @Param to_time query string false "End time (RFC3339)" Format(date-time)
// @Param last_lines query integer false "Number of recent lines" minimum(1)
// @Param last_minutes query integer false "Number of recent minutes" minimum(1)
// @Param format query string false "Output format (json, jsonpretty, csv, text)" Enums(json,jsonpretty,csv,text) default(json)
// @Success 200 {object} LogResponse
// @Failure 400 {string} string "Invalid parameters"
// @Failure 401 {string} string "Unauthorized"
// @Failure 405 {string} string "Method not allowed"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/logging/log [get]
func (h *HTTPHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	var req LogRequest

	switch r.Method {
	case http.MethodGet:
		// Parse query parameters
		fromTimeStr := r.URL.Query().Get("from_time")
		if fromTimeStr != "" {
			fromTime, err := time.Parse(time.RFC3339, fromTimeStr)
			if err != nil {
				http.Error(w, "Invalid from_time format. Use RFC3339", http.StatusBadRequest)
				return
			}
			req.FromTime = &fromTime
		}

		toTimeStr := r.URL.Query().Get("to_time")
		if toTimeStr != "" {
			toTime, err := time.Parse(time.RFC3339, toTimeStr)
			if err != nil {
				http.Error(w, "Invalid to_time format. Use RFC3339", http.StatusBadRequest)
				return
			}
			req.ToTime = &toTime
		}

		lastLinesStr := r.URL.Query().Get("last_lines")
		if lastLinesStr != "" {
			var lastLines int
			if _, err := fmt.Sscanf(lastLinesStr, "%d", &lastLines); err != nil {
				http.Error(w, "Invalid last_lines format. Must be a number", http.StatusBadRequest)
				return
			}
			req.LastLines = &lastLines
		}

		lastMinutesStr := r.URL.Query().Get("last_minutes")
		if lastMinutesStr != "" {
			var lastMinutes int
			if _, err := fmt.Sscanf(lastMinutesStr, "%d", &lastMinutes); err != nil {
				http.Error(w, "Invalid last_minutes format. Must be a number", http.StatusBadRequest)
				return
			}
			req.LastMinutes = &lastMinutes
		}

		req.Format = r.URL.Query().Get("format")

	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate format
	if req.Format == "" {
		req.Format = "json" // Default format
	} else {
		switch req.Format {
		case "json", "jsonpretty", "csv", "text":
			// Valid format
		default:
			http.Error(w, "Invalid format. Must be one of: json, jsonpretty, csv, text", http.StatusBadRequest)
			return
		}
	}

	// Handle lastMinutes parameter
	if req.LastMinutes != nil {
		now := time.Now()
		fromTime := now.Add(time.Duration(-*req.LastMinutes) * time.Minute)
		req.FromTime = &fromTime
		req.ToTime = &now
	}

	// Set default values if needed
	if req.LastLines == nil && req.FromTime == nil && req.ToTime == nil {
		defaultLines := 100
		req.LastLines = &defaultLines
	}

	// If ToTime is provided without FromTime, set FromTime to 1 hour before
	if req.FromTime == nil && req.ToTime != nil {
		fromTime := req.ToTime.Add(-1 * time.Hour)
		req.FromTime = &fromTime
	}

	// Get the log file path from the logger
	logFile := h.logger.GetLogFile()
	if logFile == "" {
		http.Error(w, "Log file path not available", http.StatusInternalServerError)
		return
	}

	// Open and read the log file
	file, err := os.Open(logFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open log file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// If we only need last N lines and no time filtering is requested
	if req.LastLines != nil && req.FromTime == nil {
		// Use a circular buffer to keep last N lines
		buffer := make([]string, 0, *req.LastLines)
		for scanner.Scan() {
			buffer = append(buffer, scanner.Text())
			if len(buffer) > *req.LastLines {
				buffer = buffer[1:]
			}
		}
		lines = buffer
	} else {
		// Time-based filtering
		for scanner.Scan() {
			line := scanner.Text()
			timestamp, err := extractTimestamp(line)
			if err != nil {
				continue // Skip lines without valid timestamp
			}

			// Check if line is within time range
			if req.FromTime != nil && timestamp.Before(*req.FromTime) {
				continue
			}
			if req.ToTime != nil && timestamp.After(*req.ToTime) {
				continue
			}

			lines = append(lines, line)
		}
	}

	if scanner.Err() != nil {
		http.Error(w, fmt.Sprintf("Error reading log file: %v", scanner.Err()), http.StatusInternalServerError)
		return
	}

	// Format and return the response based on requested format
	switch req.Format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LogResponse{Lines: lines})

	case "jsonpretty":
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(LogResponse{Lines: lines})

	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=logs.csv")
		writer := csv.NewWriter(w)
		// Write header
		writer.Write([]string{"Timestamp", "Level", "Message"})
		// Write log entries
		for _, line := range lines {
			parts := strings.SplitN(line, " ", 4)
			if len(parts) >= 4 {
				timestamp := parts[0] + " " + parts[1]
				level := strings.Trim(parts[2], "[]")
				message := parts[3]
				writer.Write([]string{timestamp, level, message})
			}
		}
		writer.Flush()

	case "text":
		w.Header().Set("Content-Type", "text/plain")
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
	}
}

// extractTimestamp attempts to parse the timestamp from a log line
func extractTimestamp(line string) (time.Time, error) {
	// Example log lines:
	// "2024/03/09 10:32:30 [INFO] Starting server..."
	// "10:32:30 [INFO] Starting server..."
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid log line format")
	}

	// Try to parse as full timestamp first
	fullTimestamp := parts[0] + " " + parts[1]
	if timestamp, err := time.Parse("2006/01/02 15:04:05", fullTimestamp); err == nil {
		return timestamp, nil
	}

	// If that fails, try to parse just the time part using today's date
	if timestamp, err := time.Parse("15:04:05", parts[0]); err == nil {
		now := time.Now()
		return time.Date(
			now.Year(), now.Month(), now.Day(),
			timestamp.Hour(), timestamp.Minute(), timestamp.Second(),
			0, time.Local,
		), nil
	}

	return time.Time{}, fmt.Errorf("invalid timestamp format: must be either '2006/01/02 15:04:05' or '15:04:05'")
}

func (h *HTTPHandler) PutWebook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PutWebook")
	fmt.Println(r.Method)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	fmt.Println("Body", string(body))
}
