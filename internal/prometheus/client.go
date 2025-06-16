package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client represents a Prometheus client for querying metrics
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// Config holds Prometheus client configuration
type Config struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// QueryResult represents the result of a Prometheus query
type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// Result represents a single result from Prometheus
type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
	Values [][]interface{}   `json:"values,omitempty"`
}

// Sample represents a time-series sample
type Sample struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// NewClient creates a new Prometheus client
func NewClient(config Config) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.URL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Query executes a PromQL query and returns the results
func (c *Client) Query(ctx context.Context, query string, timestamp time.Time) (*QueryResult, error) {
	params := url.Values{}
	params.Set("query", query)
	if !timestamp.IsZero() {
		params.Set("time", strconv.FormatInt(timestamp.Unix(), 10))
	}

	url := fmt.Sprintf("%s/api/v1/query?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s (%s)", result.Error, result.ErrorType)
	}

	return &result, nil
}

// QueryRange executes a range query and returns the results
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", fmt.Sprintf("%.0fs", step.Seconds()))

	url := fmt.Sprintf("%s/api/v1/query_range?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus range query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus range query failed: %s (%s)", result.Error, result.ErrorType)
	}

	return &result, nil
}

// GetLLDPNeighbors retrieves LLDP neighbor information from Prometheus
func (c *Client) GetLLDPNeighbors(ctx context.Context) (*QueryResult, error) {
	// Query for LLDP neighbor information
	// This assumes LLDP metrics are exposed with labels like:
	// lldp_neighbor_info{instance="device1", local_port="eth0", remote_chassis_id="...", remote_port_id="..."}
	query := `lldp_neighbor_info`

	return c.Query(ctx, query, time.Time{})
}

// GetDeviceInfo retrieves device information from Prometheus
func (c *Client) GetDeviceInfo(ctx context.Context) (*QueryResult, error) {
	// Query for device information from SNMP or other sources
	// This could include device type, model, location, etc.
	query := `{__name__=~"device_info|snmp_device_info"}`

	return c.Query(ctx, query, time.Time{})
}

// GetInterfaceInfo retrieves interface information from Prometheus
func (c *Client) GetInterfaceInfo(ctx context.Context) (*QueryResult, error) {
	// Query for interface information
	query := `{__name__=~"interface_info|snmp_if_.*"}`

	return c.Query(ctx, query, time.Time{})
}

// Health checks if Prometheus is healthy and reachable
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/-/healthy", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prometheus health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// ParseSamples converts QueryResult to a slice of Sample structs
func (c *Client) ParseSamples(result *QueryResult) ([]Sample, error) {
	var samples []Sample

	for _, r := range result.Data.Result {
		// Handle instant vector results
		if len(r.Value) == 2 {
			timestamp, ok := r.Value[0].(float64)
			if !ok {
				continue
			}

			valueStr, ok := r.Value[1].(string)
			if !ok {
				continue
			}

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			sample := Sample{
				Timestamp: time.Unix(int64(timestamp), 0),
				Value:     value,
				Labels:    make(map[string]string),
			}

			// Copy labels
			for k, v := range r.Metric {
				sample.Labels[k] = v
			}

			samples = append(samples, sample)
		}

		// Handle range vector results
		if len(r.Values) > 0 {
			for _, valueRow := range r.Values {
				if len(valueRow) != 2 {
					continue
				}

				timestamp, ok := valueRow[0].(float64)
				if !ok {
					continue
				}

				valueStr, ok := valueRow[1].(string)
				if !ok {
					continue
				}

				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					continue
				}

				sample := Sample{
					Timestamp: time.Unix(int64(timestamp), 0),
					Value:     value,
					Labels:    make(map[string]string),
				}

				// Copy labels
				for k, v := range r.Metric {
					sample.Labels[k] = v
				}

				samples = append(samples, sample)
			}
		}
	}

	return samples, nil
}
