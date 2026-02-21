package makecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a minimal Make.com API client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient creates a Client for the given zone (e.g. "eu1", "us1") and API token.
func NewClient(zone, token string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("https://%s.make.com/api/v2", zone),
		token:   token,
		http:    &http.Client{},
	}
}

// CreateScenario serialises the blueprint and scheduling, then calls POST /scenarios.
func (c *Client) CreateScenario(teamID int, bp Blueprint, sched Scheduling, folderID int) (ScenarioMeta, error) {
	bpBytes, err := json.Marshal(bp)
	if err != nil {
		return ScenarioMeta{}, fmt.Errorf("marshal blueprint: %w", err)
	}
	schedBytes, err := json.Marshal(sched)
	if err != nil {
		return ScenarioMeta{}, fmt.Errorf("marshal scheduling: %w", err)
	}
	payload := CreateScenarioRequest{
		Blueprint:  string(bpBytes),
		TeamID:     teamID,
		Scheduling: string(schedBytes),
		FolderID:   folderID,
	}
	var resp CreateScenarioResponse
	if err := c.do(http.MethodPost, "/scenarios", payload, &resp); err != nil {
		return ScenarioMeta{}, err
	}
	return resp.Scenario, nil
}


func (c *Client) do(method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, respBytes)
	}

	if out != nil {
		return json.Unmarshal(respBytes, out)
	}
	return nil
}
