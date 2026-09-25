package azure

import (
	"encoding/json"
	"testing"
)

func TestNewEstimatePayload(t *testing.T) {
	p := NewEstimatePayload("Test Architecture", "BRL")

	if p.Name != "Test Architecture" {
		t.Errorf("expected name 'Test Architecture', got %v", p.Name)
	}
	if p.Currency != "brl" {
		t.Errorf("expected currency 'brl', got %v", p.Currency)
	}
	if p.PriceLevelID != "mca" {
		t.Errorf("CRITICAL: PriceLevelID MUST be 'mca' to avoid login modal popup; got %v", p.PriceLevelID)
	}
	if p.Support.Level != "none" {
		t.Errorf("expected support level 'none', got %v", p.Support.Level)
	}
	if p.Instances == nil {
		t.Errorf("expected non-nil instances slice")
	}
}

func TestTokenRegexExtraction(t *testing.T) {
	sampleHTML1 := `<html><body><input name="__RequestVerificationToken" type="hidden" value="abc123tokenXYZ" /></body></html>`
	sampleHTML2 := `<html><body><input type="hidden" value="my_secret_token_456" name="__RequestVerificationToken" /></body></html>`

	m1 := tokenRegex1.FindStringSubmatch(sampleHTML1)
	if len(m1) < 2 || m1[1] != "abc123tokenXYZ" {
		t.Errorf("tokenRegex1 failed on sampleHTML1, got %v", m1)
	}

	m2 := tokenRegex2.FindStringSubmatch(sampleHTML2)
	if len(m2) < 2 || m2[1] != "my_secret_token_456" {
		t.Errorf("tokenRegex2 failed on sampleHTML2, got %v", m2)
	}
}

func TestPayloadJSONSerialization(t *testing.T) {
	p := NewEstimatePayload("Enterprise TIM v1", "BRL")
	vm := BuildInstance(ResourceInput{
		Service: "vm",
		Name:    "App VM",
		SKU:     "D2s_v5",
	})
	p.Instances = append(p.Instances, vm)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal generated JSON: %v", err)
	}

	if parsed["priceLevelId"] != "mca" {
		t.Errorf("expected JSON priceLevelId 'mca', got %v", parsed["priceLevelId"])
	}
	if parsed["currency"] != "brl" {
		t.Errorf("expected JSON currency 'brl', got %v", parsed["currency"])
	}

	instances, ok := parsed["instances"].([]interface{})
	if !ok || len(instances) != 1 {
		t.Fatalf("expected 1 instance in JSON, got %v", instances)
	}

	instMap, ok := instances[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected instance to be map, got %T", instances[0])
	}

	if instMap["hasPricing"] != true {
		t.Errorf("expected instance hasPricing true in JSON, got %v", instMap["hasPricing"])
	}
	if instMap["region"] != "brazil-south" {
		t.Errorf("expected instance region 'brazil-south', got %v", instMap["region"])
	}
}
