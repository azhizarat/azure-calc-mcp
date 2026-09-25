package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"azure-calc-mcp/azure"
)

func TestFormatEstimateResponse(t *testing.T) {
	result := &azure.EstimateResult{
		ID:            "abc-123",
		Hash:          "abc123",
		ShareURL:      "https://azure.com/e/abc123",
		CalculatorURL: "https://azure.microsoft.com/en-us/pricing/calculator/?shared-estimate=abc123",
		Name:          "TIM DEV v1",
		Currency:      "BRL",
		ItemCount:     2,
	}

	opts := azure.RequestOptions{
		Name:     "TIM DEV v1",
		Currency: "BRL",
		Resources: []azure.ResourceInput{
			{
				Service:  "aks",
				Name:     "AKS Cluster",
				Region:   "brazil-south",
				Quantity: 2,
				SKU:      "d2sv5",
			},
			{
				Service:   "postgresql",
				Name:      "Postgres DB",
				Region:    "brazil-south",
				StorageGB: 32,
			},
		},
	}

	resp := formatEstimateResponse(result, opts)

	if !strings.Contains(resp, "https://azure.com/e/abc123") {
		t.Errorf("expected share URL in response, got %s", resp)
	}
	if !strings.Contains(resp, "TIM DEV v1") {
		t.Errorf("expected estimate name in response, got %s", resp)
	}
	if !strings.Contains(resp, "AKS Cluster") {
		t.Errorf("expected AKS Cluster in response, got %s", resp)
	}
	if !strings.Contains(resp, "Postgres DB") {
		t.Errorf("expected Postgres DB in response, got %s", resp)
	}
	if !strings.Contains(resp, "BRL") {
		t.Errorf("expected currency BRL in response, got %s", resp)
	}
}

func TestJSONRPCSerialization(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal JSONRPCResponse: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got %v", parsed["jsonrpc"])
	}
	if parsed["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", parsed["id"])
	}
}
