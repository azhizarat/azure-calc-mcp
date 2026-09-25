package azure

import (
	"strings"
	"time"
)

// EstimatePayload represents the JSON body sent to /api/v2/calculator/shared-estimates/save/
type EstimatePayload struct {
	ID                  *string                  `json:"id"`
	EstimateTimestamp   string                   `json:"estimateTimestamp"`
	Instances           []map[string]interface{} `json:"instances"`
	Name                string                   `json:"name"`
	EnrollmentID        *string                  `json:"enrollmentId"`
	PriceLevelID        string                   `json:"priceLevelId"`
	AdditionalDiscounts []string                 `json:"additionalDiscounts"`
	MarkupValue         float64                  `json:"markupValue"`
	Total               float64                  `json:"total"`
	Support             SupportConfig            `json:"support"`
	Currency            string                   `json:"currency"`
}

// SupportConfig represents the support tier configuration
type SupportConfig struct {
	Level string  `json:"level"`
	Total float64 `json:"total"`
}

// EstimateResult contains the generated share links and metadata
type EstimateResult struct {
	ID            string `json:"id"`
	Hash          string `json:"hash"`
	ShareURL      string `json:"shareUrl"`
	CalculatorURL string `json:"calculatorUrl"`
	Name          string `json:"name"`
	Currency      string `json:"currency"`
	ItemCount     int    `json:"itemCount"`
}

// ResourceInput is a flexible specification of a single Azure service to add
type ResourceInput struct {
	Service       string                 `json:"service"`                 // "vm", "virtual-machines", "sql-database", "storage", "app-service", etc.
	Name          string                 `json:"name,omitempty"`          // Friendly name for this instance
	Region        string                 `json:"region,omitempty"`        // e.g. "eastus", "brazilsouth", "westeurope"
	Quantity      int                    `json:"quantity,omitempty"`      // e.g. 1, 2, 5
	Hours         int                    `json:"hours,omitempty"`         // monthly hours, default 730
	BillingOption string                 `json:"billingOption,omitempty"` // "one-hour", "one-year", "three-year", "savings-plan"

	// VM specific
	OS       string `json:"os,omitempty"`       // "linux" (default) or "windows"
	SKU      string `json:"sku,omitempty"`      // e.g. "D4s_v5", "B2s", "Standard_D8s_v5"
	DiskTier string `json:"diskTier,omitempty"` // "premiumssd", "standardssd", "standardhdd"
	DiskSize string `json:"diskSize,omitempty"` // "P10", "P15", "P20", "P30", "128GB", "256GB", etc.

	// SQL specific
	Tier        string `json:"tier,omitempty"`        // "general-purpose", "business-critical", "hyperscale"
	ComputeTier string `json:"computeTier,omitempty"` // "provisioned" (default) or "serverless"
	VCores      int    `json:"vcores,omitempty"`      // 2, 4, 8, 16, etc.
	StorageGB   int    `json:"storageGb,omitempty"`   // Storage in GB

	// Storage Account (Blob) specific
	Redundancy string `json:"redundancy,omitempty"` // "lrs", "zrs", "grs"
	AccessTier string `json:"accessTier,omitempty"` // "hot", "cool", "cold", "archive"

	// App Service specific
	AppServicePlan string `json:"appServicePlan,omitempty"` // "B1", "S1", "P1v3", "P2v3", etc.

	// AKS specific
	ClusterCount int    `json:"clusterCount,omitempty"` // Number of clusters (default 1)
	SLAOption    string `json:"slaOption,omitempty"`    // "no-sla-free-non-production" or "sla"

	// PostgreSQL specific
	DeploymentType string `json:"deploymentType,omitempty"` // "flexibleserver"
	ComputeType    string `json:"computeType,omitempty"`    // e.g. "flexible-server-burstable-compute-b1ms"

	// Key Vault specific
	Operations float64 `json:"operations,omitempty"` // Number of 10,000 operation units

	// Azure Monitor specific
	DailyLogsIngested float64 `json:"dailyLogsIngested,omitempty"` // GB per day

	// Raw / passthrough parameters if needed
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// RequestOptions defines options for creating an estimate
type RequestOptions struct {
	Name      string          `json:"name"`
	Currency  string          `json:"currency"` // "usd", "brl", "eur", etc.
	Resources []ResourceInput `json:"resources"`
}

// NewEstimatePayload initializes a payload with default values
func NewEstimatePayload(name, currency string) *EstimatePayload {
	if currency == "" {
		currency = "brl"
	}
	if name == "" {
		name = "Azure Estimate"
	}
	return &EstimatePayload{
		ID:                  nil,
		EstimateTimestamp:   time.Now().UTC().Format(time.RFC3339),
		Instances:           make([]map[string]interface{}, 0),
		Name:                name,
		EnrollmentID:        nil,
		PriceLevelID:        "mca", // "mca" (Microsoft Customer Agreement) avoids the MOSA login prompt modal
		AdditionalDiscounts: []string{},
		MarkupValue:         0,
		Total:               0,
		Support: SupportConfig{
			Level: "none",
			Total: 0,
		},
		Currency: strings.ToLower(currency),
	}
}
