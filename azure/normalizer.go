package azure

import (
	"strings"
)

// NormalizeRegionSlug ensures regions match the kebab-case slugs required by Azure Pricing Calculator
func NormalizeRegionSlug(region string) string {
	r := strings.ToLower(strings.TrimSpace(region))
	r = strings.ReplaceAll(r, " ", "")
	r = strings.ReplaceAll(r, "_", "-")

	mapping := map[string]string{
		"brazilsouth":        "brazil-south",
		"brazilsoutheast":    "brazil-southeast",
		"eastus":             "us-east",
		"eastus2":            "us-east-2",
		"westus":             "us-west",
		"westus2":            "us-west-2",
		"westus3":            "us-west-3",
		"centralus":          "us-central",
		"northcentralus":     "us-north-central",
		"southcentralus":     "us-south-central",
		"westcentralus":      "us-west-central",
		"westeurope":         "europe-west",
		"northeurope":        "europe-north",
		"uksouth":            "uk-south",
		"ukwest":             "uk-west",
		"japaneast":          "japan-east",
		"japanwest":          "japan-west",
		"southeastasia":      "asia-southeast",
		"eastasia":           "asia-east",
		"australiaeast":      "australia-east",
		"australiasoutheast": "australia-southeast",
		"australiacentral":   "australia-central",
		"australiacentral2":  "australia-central-2",
		"centralindia":       "india-central",
		"southindia":         "india-south",
		"westindia":          "india-west",
		"canadacentral":      "canada-central",
		"canadaeast":         "canada-east",
		"germanywestcentral": "germany-west-central",
		"germanynorth":       "germany-north",
		"francecentral":      "france-central",
		"francesouth":        "france-south",
		"norwayeast":         "norway-east",
		"norwaywest":         "norway-west",
		"switzerlandnorth":   "switzerland-north",
		"switzerlandwest":    "switzerland-west",
		"swedencentral":      "sweden-central",
		"swedensouth":        "sweden-south",
		"polandcentral":      "poland-central",
		"italynorth":         "italy-north",
		"spaincentral":       "spain-central",
		"mexicocentral":      "mexico-central",
		"uaenorth":           "uae-north",
		"uaecentral":         "uae-central",
		"southafricanorth":   "south-africa-north",
		"southafricawest":    "south-africa-west",
		"koreacentral":       "koreacentral",
		"koreasouth":         "koreasouth",
	}

	if val, ok := mapping[r]; ok {
		return val
	}
	if r == "" {
		return "brazil-south"
	}
	return r
}

// NormalizeServiceSlug normalizes any friendly name or common alias to the official Azure Calculator slug
func NormalizeServiceSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	switch s {
	case "vm", "vms", "virtual-machine", "virtual-machines", "virtualmachines", "compute":
		return "virtual-machines"
	case "sql", "sql-database", "sqldatabase", "azure-sql", "azure-sql-database":
		return "azure-sql-database"
	case "app-service", "appservice", "webapp", "web-app", "web-apps":
		return "app-service"
	case "storage", "blob", "blob-storage", "storage-account", "storage-accounts":
		return "storage"
	case "aks", "kubernetes", "kubernetes-service", "azure-kubernetes-service":
		return "kubernetes-service"
	case "postgres", "postgresql", "azure-database-for-postgresql", "flexible-server", "postgres-flexible":
		return "postgresql"
	case "mysql", "azure-database-for-mysql", "mysql-flexible":
		return "mysql"
	case "cosmos", "cosmos-db", "cosmosdb", "azure-cosmos-db":
		return "cosmos-db"
	case "openai", "azure-openai", "cognitive-services", "ai-studio", "azure-ai":
		return "cognitive-services"
	case "key-vault", "keyvault", "vault":
		return "key-vault"
	case "monitor", "azure-monitor", "log-analytics", "logs":
		return "monitor"
	case "app-gateway", "application-gateway", "appgateway":
		return "application-gateway"
	case "vpn", "vpn-gateway", "vpngateway":
		return "vpn-gateway"
	case "firewall", "azure-firewall":
		return "azure-firewall"
	case "container-apps", "containerapps", "aca":
		return "container-apps"
	case "functions", "azure-functions", "serverless-functions":
		return "functions"
	case "redis", "redis-cache", "azure-cache-for-redis":
		return "redis-cache"
	case "service-bus", "servicebus":
		return "service-bus"
	case "event-hubs", "eventhubs", "event-hub":
		return "event-hubs"
	case "api-management", "apim":
		return "api-management"
	case "bastion", "azure-bastion":
		return "azure-bastion"
	default:
		return s
	}
}

// NormalizeVMSize formats SKU to Azure calculator standard
func NormalizeVMSize(sku string) string {
	sku = strings.TrimSpace(strings.ToLower(sku))
	sku = strings.TrimPrefix(sku, "standard_")
	sku = strings.TrimPrefix(sku, "basic_")
	sku = strings.ReplaceAll(sku, "_", "-")
	return sku
}

// NormalizeDiskSize formats disk sizes to Azure disk tiers (p10, e10, etc.)
func NormalizeDiskSize(size string) string {
	size = strings.TrimSpace(strings.ToLower(size))
	size = strings.TrimSuffix(size, "gb")
	size = strings.TrimSuffix(size, "gib")

	switch size {
	case "32", "p4":
		return "p4"
	case "64", "p6":
		return "p6"
	case "128", "p10":
		return "p10"
	case "256", "p15":
		return "p15"
	case "512", "p20":
		return "p20"
	case "1024", "1000", "1tb", "p30":
		return "p30"
	case "2048", "2000", "2tb", "p40":
		return "p40"
	case "4096", "4000", "4tb", "p50":
		return "p50"
	default:
		if strings.HasPrefix(size, "p") || strings.HasPrefix(size, "e") {
			return size
		}
		return "p15"
	}
}
