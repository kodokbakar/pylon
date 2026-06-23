//go:build e2e

package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type healthEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		Service string `json:"service"`
		Status  string `json:"status"`
	} `json:"data"`
}

func TestE2EHealthAndMetrics(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("PYLON_E2E_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("PYLON_E2E_BASE_URL is not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	healthResp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("get health: %v", err)
	}
	defer healthResp.Body.Close()

	if healthResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(healthResp.Body)
		t.Fatalf("expected health status 200, got %d body=%s", healthResp.StatusCode, body)
	}

	var health healthEnvelope
	if err := json.NewDecoder(healthResp.Body).Decode(&health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if !health.Success || health.Data.Service != "api-gateway" || health.Data.Status != "ok" {
		t.Fatalf("unexpected health response: %#v", health)
	}

	metricsResp, err := client.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("get metrics: %v", err)
	}
	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(metricsResp.Body)
		t.Fatalf("expected metrics status 200, got %d body=%s", metricsResp.StatusCode, body)
	}

	metricsBody, err := io.ReadAll(metricsResp.Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}

	if !strings.Contains(string(metricsBody), "pylon_http_requests_total") {
		t.Fatalf("expected metrics to contain pylon_http_requests_total")
	}
}
