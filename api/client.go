package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const baseURL = "https://neocities.org/api"

// Client handles communication with the Neocities API.
type Client struct {
	apiKey     string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new API client with an API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// NewClientWithCredentials creates a new API client with username/password.
func NewClientWithCredentials(username, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		httpClient: &http.Client{},
	}
}

// FileInfo represents a file in the Neocities site.
type FileInfo struct {
	Path      string `json:"path"`
	IsDir     bool   `json:"is_directory"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
	SHA1Hash  string `json:"sha1_hash"`
}

// ListResponse represents the response from the list endpoint.
type ListResponse struct {
	Result string     `json:"result"`
	Files  []FileInfo `json:"files"`
}

// SiteInfo represents information about a Neocities site.
type SiteInfo struct {
	Sitename  string   `json:"sitename"`
	Hits      int64    `json:"hits"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"last_updated"`
	Domain    string   `json:"domain"`
	Tags      []string `json:"tags"`
}

// InfoResponse represents the response from the info endpoint.
type InfoResponse struct {
	Result string   `json:"result"`
	Info   SiteInfo `json:"info"`
}

// KeyResponse represents the response from the key endpoint.
type KeyResponse struct {
	Result string `json:"result"`
	APIKey string `json:"api_key"`
}

// APIResponse represents a generic API response.
type APIResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	return c.httpClient.Do(req)
}

// Key retrieves the API key using username/password authentication.
func (c *Client) Key() (string, error) {
	req, err := http.NewRequest("GET", baseURL+"/key", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var keyResp KeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&keyResp); err != nil {
		return "", err
	}

	if keyResp.Result != "success" {
		return "", fmt.Errorf("API returned non-success result")
	}

	return keyResp.APIKey, nil
}

// List retrieves the file listing for the site.
func (c *Client) List(path string) ([]FileInfo, error) {
	url := baseURL + "/list"
	if path != "" {
		url += "?path=" + path
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	if listResp.Result != "success" {
		return nil, fmt.Errorf("API returned non-success result")
	}

	return listResp.Files, nil
}

// Info retrieves information about a site.
func (c *Client) Info(sitename string) (*SiteInfo, error) {
	url := baseURL + "/info"
	if sitename != "" {
		url += "?sitename=" + sitename
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var infoResp InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&infoResp); err != nil {
		return nil, err
	}

	if infoResp.Result != "success" {
		return nil, fmt.Errorf("API returned non-success result")
	}

	return &infoResp.Info, nil
}

// Upload uploads files to the Neocities site.
// files is a map of remote path -> local file path.
func (c *Client) Upload(files map[string]string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for remotePath, localPath := range files {
		file, err := os.Open(localPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", localPath, err)
		}

		part, err := writer.CreateFormFile(remotePath, filepath.Base(localPath))
		if err != nil {
			file.Close()
			return err
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/upload", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if apiResp.Result != "success" {
		return fmt.Errorf("upload failed: %s", apiResp.Message)
	}

	return nil
}

// Delete removes files from the Neocities site.
func (c *Client) Delete(filenames []string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for _, filename := range filenames {
		if err := writer.WriteField("filenames[]", filename); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/delete", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if apiResp.Result != "success" {
		return fmt.Errorf("delete failed: %s", apiResp.Message)
	}

	return nil
}
