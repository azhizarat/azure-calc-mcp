package azure

import (
	"strings"
)

// BuildInstance converts a friendly ResourceInput into an Azure Pricing Calculator instance map
func BuildInstance(input ResourceInput) map[string]interface{} {
	slug := NormalizeServiceSlug(input.Service)
	region := NormalizeRegionSlug(input.Region)

	instanceName := input.Name
	if instanceName == "" {
		instanceName = formatDefaultName(slug)
	}

	count := input.Quantity
	if count <= 0 {
		count = 1
	}

	hours := input.Hours
	if hours <= 0 {
		hours = 730
	}

	billingOption := input.BillingOption
	if billingOption == "" {
		billingOption = "one-hour"
	}

	instance := GetDefaultSchema(slug)
	if instance == nil {
		instance = make(map[string]interface{})
	}

	// Always ensure serviceSlug, displayName, region, and hasPricing: true
	instance["serviceSlug"] = slug
	instance["displayName"] = instanceName
	instance["region"] = region
	instance["hasPricing"] = true

	switch slug {
	case "virtual-machines":
		buildVMInstance(instance, input, count, hours, billingOption)
	case "azure-sql-database":
		buildSQLInstance(instance, input, count, hours, billingOption)
	case "app-service":
		buildAppServiceInstance(instance, input, count, hours, billingOption)
	case "storage":
		buildStorageInstance(instance, input)
	case "kubernetes-service":
		buildAKSInstance(instance, input, count, hours)
	case "postgresql":
		buildPostgresInstance(instance, input, count, hours)
	case "key-vault":
		buildKeyVaultInstance(instance, input)
	case "monitor":
		buildMonitorInstance(instance, input)
	}

	// Merge any extra parameters provided by the caller
	for k, v := range input.Extra {
		instance[k] = v
	}

	return instance
}

func buildVMInstance(m map[string]interface{}, input ResourceInput, count, hours int, billingOption string) {
	os := strings.ToLower(input.OS)
	if os == "" || !strings.Contains(os, "win") {
		os = "linux"
	} else {
		os = "windows"
	}

	size := NormalizeVMSize(input.SKU)
	if size == "" {
		size = "d4s-v5"
	}

	diskTier := strings.ToLower(input.DiskTier)
	if diskTier == "" || strings.Contains(diskTier, "premium") {
		diskTier = "premiumssd"
	} else if strings.Contains(diskTier, "standard") && strings.Contains(diskTier, "hdd") {
		diskTier = "standardhdd"
	} else if strings.Contains(diskTier, "standard") {
		diskTier = "standardssd"
	}

	diskSize := NormalizeDiskSize(input.DiskSize)
	if diskSize == "" {
		diskSize = "p15"
	}

	m["operatingSystem"] = os
	m["type"] = "os-only"
	m["tier"] = "standard"
	m["size"] = size
	m["count"] = count
	m["hours"] = hours
	m["hoursFactor"] = 1
	m["computeBillingOption"] = billingOption
	m["managedDiskTier"] = diskTier
	m["managedDisks"] = diskSize
	m["license"] = "none"
}

func buildSQLInstance(m map[string]interface{}, input ResourceInput, count, hours int, billingOption string) {
	vcoreTier := strings.ToLower(input.Tier)
	if vcoreTier == "" || strings.Contains(vcoreTier, "general") {
		vcoreTier = "general-purpose"
	} else if strings.Contains(vcoreTier, "business") {
		vcoreTier = "business-critical"
	} else if strings.Contains(vcoreTier, "hyper") {
		vcoreTier = "hyperscale"
	}

	computeTier := strings.ToLower(input.ComputeTier)
	if computeTier == "" || strings.Contains(computeTier, "prov") {
		computeTier = "provisioned"
	} else {
		computeTier = "serverless"
	}

	vcores := input.VCores
	if vcores <= 0 {
		vcores = 4
	}

	storageGB := input.StorageGB
	if storageGB <= 0 {
		storageGB = 32
	}

	m["type"] = "single"
	m["tier"] = "vcore"
	m["vcoreTier"] = vcoreTier
	m["computeTier"] = computeTier
	m["vcores"] = vcores
	m["storage"] = storageGB
	m["storageFactor"] = 1
	m["hours"] = hours
	m["hoursFactor"] = 1
}

func buildAppServiceInstance(m map[string]interface{}, input ResourceInput, count, hours int, billingOption string) {
	os := strings.ToLower(input.OS)
	if os == "" || !strings.Contains(os, "win") {
		os = "linux"
	} else {
		os = "windows"
	}

	plan := strings.ToLower(input.AppServicePlan)
	tier := "premium-v3"
	size := "p1v3"

	if plan != "" {
		if strings.HasPrefix(plan, "p") {
			tier = "premium-v3"
			size = plan
		} else if strings.HasPrefix(plan, "s") {
			tier = "standard"
			size = plan
		} else if strings.HasPrefix(plan, "b") {
			tier = "basic"
			size = plan
		}
	}

	m["type"] = os
	m["tier"] = tier
	m["size"] = size
	m["instances"] = count
	m["hours"] = hours
	m["hoursFactor"] = 1
	m["billingOption"] = billingOption
}

func buildStorageInstance(m map[string]interface{}, input ResourceInput) {
	redundancy := strings.ToLower(input.Redundancy)
	if redundancy == "" {
		redundancy = "lrs"
	}

	accessTier := strings.ToLower(input.AccessTier)
	if accessTier == "" {
		accessTier = "hot"
	}

	capacityGB := input.StorageGB
	if capacityGB <= 0 {
		capacityGB = 20
	}

	m["storageAccountType"] = "general-purpose-v2"
	m["type"] = "block-blob"
	m["performanceTier"] = "standard"
	m["tier"] = "standard"
	m["fileStructure"] = "flat"
	m["accessTier"] = accessTier
	m["redundancy"] = redundancy
	m["blobStorage"] = capacityGB
	m["blobStorageFactor"] = 1
	m["count"] = float64(capacityGB)
	m["dataAtRestUnits"] = float64(capacityGB)
	m["blobWriteOperations"] = 10.0
	m["blobCreateContainerOperations"] = 10.0
	m["blobReadOperations"] = 10.0
	m["blobOtherOperations"] = 1
	m["blobDataRetrieval"] = 1000.0
	m["blobDataWrite"] = 1000.0
	m["sftpToggle"] = false
	m["blobObjectReplicationToggle"] = false
}

func buildAKSInstance(m map[string]interface{}, input ResourceInput, count, hours int) {
	size := strings.ToLower(input.SKU)
	if size == "" {
		size = "d2sv5"
	}
	size = strings.ReplaceAll(size, "-", "")
	size = strings.ReplaceAll(size, "_", "")

	clusterCount := input.ClusterCount
	if clusterCount <= 0 {
		clusterCount = 1
	}

	slaOption := input.SLAOption
	if slaOption == "" {
		slaOption = "no-sla-free-non-production"
	}

	os := strings.ToLower(input.OS)
	if os == "" || !strings.Contains(os, "win") {
		os = "linux"
	} else {
		os = "windows"
	}

	diskTier := strings.ToLower(input.DiskTier)
	if diskTier == "" {
		diskTier = "standardssd"
	}

	diskType := strings.ToLower(input.DiskSize)
	if diskType == "" {
		diskType = "e10"
	}

	m["size"] = size
	m["count"] = count
	m["hours"] = float64(hours)
	m["operatingSystem"] = os
	m["tier"] = "standard"
	m["managedDiskTier"] = diskTier
	m["managedDiskType"] = diskType
	m["managedDisks"] = count
	m["clusterCount"] = clusterCount
	m["slaOptionValue"] = slaOption
}

func buildPostgresInstance(m map[string]interface{}, input ResourceInput, count, hours int) {
	computeType := input.ComputeType
	if computeType == "" {
		computeType = "flexible-server-burstable-compute-b1ms"
	}

	storageGB := input.StorageGB
	if storageGB <= 0 {
		storageGB = 32
	}

	m["deploymentType"] = "flexibleserver"
	m["tier"] = "burstable"
	m["burstableType"] = computeType
	m["computeType"] = computeType
	m["singleServers"] = count
	m["singleHours"] = float64(hours)
	m["storageCount"] = storageGB
	m["premiumSsdTier"] = "premiumssd"
	m["extendedSupportToggler"] = false
}

func buildKeyVaultInstance(m map[string]interface{}, input ResourceInput) {
	ops := input.Operations
	if ops <= 0 {
		ops = 1.0
	}
	m["operations"] = ops
	m["advancedOperations"] = 0.0
	m["renewals"] = 0
	m["keys"] = 0
	m["advancedKeys"] = 0
	m["hsmPoolsUnits"] = 0
}

func buildMonitorInstance(m map[string]interface{}, input ResourceInput) {
	dailyLogs := input.DailyLogsIngested
	if dailyLogs <= 0 {
		dailyLogs = 0.1
	}
	m["dailyLogsIngested"] = dailyLogs
	m["logAnalyticsRetention"] = 1
}



func formatDefaultName(slug string) string {
	switch slug {
	case "virtual-machines":
		return "Virtual Machine"
	case "azure-sql-database":
		return "Azure SQL Database"
	case "app-service":
		return "App Service"
	case "storage":
		return "Storage Account"
	case "kubernetes-service":
		return "Azure Kubernetes Service (AKS)"
	case "postgresql":
		return "Azure Database for PostgreSQL"
	case "key-vault":
		return "Key Vault"
	case "monitor":
		return "Azure Monitor"
	default:
		return strings.Title(strings.ReplaceAll(slug, "-", " "))
	}
}
