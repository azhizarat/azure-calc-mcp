package azure

import (
	"testing"
)

func TestNormalizeRegionSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"brazilsouth", "brazil-south"},
		{"Brazil South", "brazil-south"},
		{"brazil_south", "brazil-south"},
		{"BRAZIL-SOUTH", "brazil-south"},
		{"eastus", "us-east"},
		{"East US", "us-east"},
		{"eastus2", "us-east-2"},
		{"westeurope", "europe-west"},
		{"northeurope", "europe-north"},
		{"uksouth", "uk-south"},
		{"japaneast", "japan-east"},
		{"australiaeast", "australia-east"},
		{"", "brazil-south"}, // default fallback
		{"already-kebab", "already-kebab"},
	}

	for _, tt := range tests {
		actual := NormalizeRegionSlug(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizeRegionSlug(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestNormalizeServiceSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"vm", "virtual-machines"},
		{"VM", "virtual-machines"},
		{"Virtual Machines", "virtual-machines"},
		{"virtual_machine", "virtual-machines"},
		{"aks", "kubernetes-service"},
		{"Kubernetes", "kubernetes-service"},
		{"Azure Kubernetes Service", "kubernetes-service"},
		{"postgres", "postgresql"},
		{"postgresql", "postgresql"},
		{"flexible-server", "postgresql"},
		{"blob", "storage"},
		{"storage-accounts", "storage"},
		{"sql", "azure-sql-database"},
		{"azure-sql", "azure-sql-database"},
		{"app-service", "app-service"},
		{"webapp", "app-service"},
		{"openai", "cognitive-services"},
		{"azure-openai", "cognitive-services"},
		{"cosmos", "cosmos-db"},
		{"cosmosdb", "cosmos-db"},
		{"key-vault", "key-vault"},
		{"vault", "key-vault"},
		{"monitor", "monitor"},
		{"log-analytics", "monitor"},
		{"app-gateway", "application-gateway"},
		{"redis", "redis-cache"},
		{"container-apps", "container-apps"},
		{"functions", "functions"},
		{"unknown-custom-service", "unknown-custom-service"},
	}

	for _, tt := range tests {
		actual := NormalizeServiceSlug(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizeServiceSlug(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestNormalizeVMSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Standard_D4s_v5", "d4s-v5"},
		{"standard_b2s", "b2s"},
		{"D2s_v5", "d2s-v5"},
		{"basic_a1", "a1"},
	}

	for _, tt := range tests {
		actual := NormalizeVMSize(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizeVMSize(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestNormalizeDiskSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"128", "p10"},
		{"128GB", "p10"},
		{"256gib", "p15"},
		{"P20", "p20"},
		{"E10", "e10"},
		{"1TB", "p30"},
	}

	for _, tt := range tests {
		actual := NormalizeDiskSize(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizeDiskSize(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}
