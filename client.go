package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Printer is a printer we report up to the back office.
type Printer struct {
	SystemName string `json:"system_name"`
	Name       string `json:"name"`
	Role       string `json:"role,omitempty"`
	IsDefault  bool   `json:"is_default,omitempty"`
}

// Job is one queued print job returned by the back office.
type Job struct {
	ID           int    `json:"id"`
	DocumentType string `json:"document_type"`
	Copies       int    `json:"copies"`
	OrderNumber  string `json:"order_number"`
	Printer      string `json:"printer"`      // OS system name to print to
	DocumentURL  string `json:"document_url"` // PDF to fetch + print
}

// Client talks to the back office print-agent API. Every call carries the device
// key in X-Agent-Key.
type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func newClient(baseURL, key string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Key:     key,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// envelope is the standard { data, error, meta } response shape.
type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// do performs a request and returns the unwrapped `data` payload.
func (c *Client) do(method, path string, body any) (json.RawMessage, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Agent-Key", c.Key)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("bad response (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if env.Error != nil {
		return nil, fmt.Errorf("%s (%s)", env.Error.Message, env.Error.Code)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return env.Data, nil
}

// Heartbeat registers/refreshes this agent's printers and liveness.
func (c *Client) Heartbeat(version, activeUser, location string, printers []Printer) error {
	_, err := c.do("POST", "/api/print/agent/heartbeat", map[string]any{
		"version":     version,
		"active_user": activeUser,
		"location":    location,
		"printers":    printers,
	})
	return err
}

// PollJobs claims and returns queued jobs for this agent's printers.
func (c *Client) PollJobs() ([]Job, error) {
	data, err := c.do("GET", "/api/print/agent/jobs", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Jobs []Job `json:"jobs"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out.Jobs, nil
}

// ReportStatus reports a job outcome (printed / failed).
func (c *Client) ReportStatus(id int, status, errMsg string) error {
	_, err := c.do("POST", fmt.Sprintf("/api/print/agent/jobs/%d/status", id), map[string]any{
		"status": status,
		"error":  errMsg,
	})
	return err
}

// download fetches a document URL to a byte slice.
func (c *Client) download(url string) ([]byte, error) {
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
