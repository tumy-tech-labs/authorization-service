package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTPClient: &http.Client{}}
}

type AccessRequest struct {
	TenantID   string            `json:"tenantID"`
	Subject    string            `json:"subject"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

type Decision struct {
	Allow    bool   `json:"allow"`
	PolicyID string `json:"policyID"`
	Reason   string `json:"reason"`
}

func (c *Client) post(path string, payload any) (*http.Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	return c.HTTPClient.Post(url, "application/json", bytes.NewReader(b))
}

func (c *Client) CheckAccess(req AccessRequest) (*Decision, error) {
	resp, err := c.post("/check-access", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	var dec Decision
	if err := json.NewDecoder(resp.Body).Decode(&dec); err != nil {
		return nil, err
	}
	return &dec, nil
}

func (c *Client) CompileRule(tenantID, rule string) (string, error) {
	resp, err := c.post("/compile", map[string]string{"tenantID": tenantID, "rule": rule})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *Client) ValidatePolicy(tenantID, policy string) error {
	resp, err := c.post("/validate-policy", map[string]string{"tenantID": tenantID, "policy": policy})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
