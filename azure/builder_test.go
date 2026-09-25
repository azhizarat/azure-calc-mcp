package azure

import (
	"testing"
)

func TestBuildVMInstance(t *testing.T) {
	input := ResourceInput{
		Service:       "vm",
		Name:          "Frontend VM",
		Region:        "brazilsouth",
		Quantity:      2,
		Hours:         730,
		OS:            "linux",
		SKU:           "Standard_D4s_v5",
		DiskTier:      "premiumssd",
		DiskSize:      "p20",
		BillingOption: "one-year",
	}

	inst := BuildInstance(input)

	if inst["serviceSlug"] != "virtual-machines" {
		t.Errorf("expected serviceSlug 'virtual-machines', got %v", inst["serviceSlug"])
	}
	if inst["region"] != "brazil-south" {
		t.Errorf("expected region 'brazil-south', got %v", inst["region"])
	}
	if inst["hasPricing"] != true {
		t.Errorf("expected hasPricing true, got %v", inst["hasPricing"])
	}
	if inst["size"] != "d4s-v5" {
		t.Errorf("expected size 'd4s-v5', got %v", inst["size"])
	}
	if inst["count"] != 2 {
		t.Errorf("expected count 2, got %v", inst["count"])
	}
	if inst["operatingSystem"] != "linux" {
		t.Errorf("expected OS 'linux', got %v", inst["operatingSystem"])
	}
	if inst["managedDiskTier"] != "premiumssd" {
		t.Errorf("expected managedDiskTier 'premiumssd', got %v", inst["managedDiskTier"])
	}
	if inst["managedDisks"] != "p20" {
		t.Errorf("expected managedDisks 'p20', got %v", inst["managedDisks"])
	}
}

func TestBuildAKSInstance(t *testing.T) {
	input := ResourceInput{
		Service:      "aks",
		Name:         "AKS - 2x D2s_v5 (Free tier)",
		Region:       "brazilsouth",
		SKU:          "d2sv5",
		Quantity:     2,
		Hours:        730,
		OS:           "linux",
		Tier:         "standard",
		DiskTier:     "standardssd",
		DiskSize:     "e10",
		ClusterCount: 1,
		SLAOption:    "no-sla-free-non-production",
	}

	inst := BuildInstance(input)

	if inst["serviceSlug"] != "kubernetes-service" {
		t.Errorf("expected serviceSlug 'kubernetes-service', got %v", inst["serviceSlug"])
	}
	if inst["region"] != "brazil-south" {
		t.Errorf("expected region 'brazil-south', got %v", inst["region"])
	}
	if inst["hasPricing"] != true {
		t.Errorf("expected hasPricing true, got %v", inst["hasPricing"])
	}
	if inst["size"] != "d2sv5" {
		t.Errorf("expected size 'd2sv5', got %v", inst["size"])
	}
	if inst["count"] != 2 {
		t.Errorf("expected count 2, got %v", inst["count"])
	}
	if inst["clusterCount"] != 1 {
		t.Errorf("expected clusterCount 1, got %v", inst["clusterCount"])
	}
	if inst["slaOptionValue"] != "no-sla-free-non-production" {
		t.Errorf("expected slaOptionValue 'no-sla-free-non-production', got %v", inst["slaOptionValue"])
	}
	if inst["managedDiskTier"] != "standardssd" {
		t.Errorf("expected managedDiskTier 'standardssd', got %v", inst["managedDiskTier"])
	}
	if inst["managedDiskType"] != "e10" {
		t.Errorf("expected managedDiskType 'e10', got %v", inst["managedDiskType"])
	}
	if inst["managedDisks"] != 2 {
		t.Errorf("expected managedDisks 2, got %v", inst["managedDisks"])
	}
}

func TestBuildPostgresInstance(t *testing.T) {
	input := ResourceInput{
		Service:        "postgresql",
		Name:           "PostgreSQL B1ms 32GB",
		Region:         "brazilsouth",
		DeploymentType: "flexibleserver",
		Tier:           "burstable",
		ComputeType:    "flexible-server-burstable-compute-b1ms",
		Quantity:       1,
		Hours:          730,
		StorageGB:      32,
	}

	inst := BuildInstance(input)

	if inst["serviceSlug"] != "postgresql" {
		t.Errorf("expected serviceSlug 'postgresql', got %v", inst["serviceSlug"])
	}
	if inst["deploymentType"] != "flexibleserver" {
		t.Errorf("expected deploymentType 'flexibleserver', got %v", inst["deploymentType"])
	}
	if inst["tier"] != "burstable" {
		t.Errorf("expected tier 'burstable', got %v", inst["tier"])
	}
	if inst["burstableType"] != "flexible-server-burstable-compute-b1ms" {
		t.Errorf("expected burstableType 'flexible-server-burstable-compute-b1ms', got %v", inst["burstableType"])
	}
	if inst["storageCount"] != 32 {
		t.Errorf("expected storageCount 32, got %v", inst["storageCount"])
	}
	if inst["hasPricing"] != true {
		t.Errorf("expected hasPricing true, got %v", inst["hasPricing"])
	}
}

func TestBuildStorageInstance(t *testing.T) {
	input := ResourceInput{
		Service:    "storage",
		Name:       "Blob 20 GB",
		Region:     "brazilsouth",
		StorageGB:  20,
		AccessTier: "hot",
		Redundancy: "lrs",
	}

	inst := BuildInstance(input)

	if inst["serviceSlug"] != "storage" {
		t.Errorf("expected serviceSlug 'storage', got %v", inst["serviceSlug"])
	}
	if inst["type"] != "block-blob" {
		t.Errorf("expected type 'block-blob', got %v", inst["type"])
	}
	if inst["accessTier"] != "hot" {
		t.Errorf("expected accessTier 'hot', got %v", inst["accessTier"])
	}
	if inst["redundancy"] != "lrs" {
		t.Errorf("expected redundancy 'lrs', got %v", inst["redundancy"])
	}
	if inst["hasPricing"] != true {
		t.Errorf("expected hasPricing true, got %v", inst["hasPricing"])
	}
	if inst["storageAccountType"] != "general-purpose-v2" {
		t.Errorf("expected storageAccountType 'general-purpose-v2', got %v", inst["storageAccountType"])
	}
}

func TestBuildGenericInstanceWithExtra(t *testing.T) {
	input := ResourceInput{
		Service: "cosmos-db",
		Name:    "Global NoSQL DB",
		Region:  "brazilsouth",
		Extra: map[string]interface{}{
			"throughput": 1000,
			"apiType":    "core",
		},
	}

	inst := BuildInstance(input)

	if inst["serviceSlug"] != "cosmos-db" {
		t.Errorf("expected serviceSlug 'cosmos-db', got %v", inst["serviceSlug"])
	}
	if inst["region"] != "brazil-south" {
		t.Errorf("expected region 'brazil-south', got %v", inst["region"])
	}
	if inst["hasPricing"] != true {
		t.Errorf("expected hasPricing true, got %v", inst["hasPricing"])
	}
	if inst["throughput"] != 1000 {
		t.Errorf("expected throughput 1000, got %v", inst["throughput"])
	}
	if inst["apiType"] != "core" {
		t.Errorf("expected apiType 'core', got %v", inst["apiType"])
	}
}
