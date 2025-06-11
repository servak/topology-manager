package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

type PrometheusClient struct {
	baseURL string
	client  *http.Client
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type LLDPMetric struct {
	Hardware       string
	IfDescr        string
	Instance       string
	LLDPRemPortId  string
	LLDPRemSysName string
}

func NewPrometheusClient() *PrometheusClient {
	baseURL := os.Getenv("PROMETHEUS_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}

	return &PrometheusClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *PrometheusClient) QueryLLDPMetrics(ctx context.Context) ([]LLDPMetric, error) {
	query := "lldpRemSysName"
	
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query", p.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := url.Values{}
	params.Add("query", query)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Prometheus query failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed: %s", promResp.Status)
	}

	var metrics []LLDPMetric
	for _, result := range promResp.Data.Result {
		metric := LLDPMetric{
			Hardware:       result.Metric["hardware"],
			IfDescr:        result.Metric["ifDescr"],
			Instance:       result.Metric["instance"],
			LLDPRemPortId:  result.Metric["lldpRemPortId"],
			LLDPRemSysName: result.Metric["lldpRemSysName"],
		}

		if metric.Instance != "" && metric.LLDPRemSysName != "" {
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func (p *PrometheusClient) Health(ctx context.Context) error {
	u := fmt.Sprintf("%s/-/healthy", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Prometheus health check failed with status %d", resp.StatusCode)
	}

	return nil
}

func extractDeviceName(instance string) string {
	re := regexp.MustCompile(`^([^.]+)\.`)
	matches := re.FindStringSubmatch(instance)
	if len(matches) > 1 {
		return matches[1]
	}
	return instance
}