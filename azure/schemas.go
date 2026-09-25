package azure

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

//go:embed default_schemas.json
var defaultSchemasJSON []byte

var (
	schemasMu      sync.RWMutex
	defaultSchemas map[string]map[string]interface{}
)

func init() {
	defaultSchemas = make(map[string]map[string]interface{})
	_ = json.Unmarshal(defaultSchemasJSON, &defaultSchemas)
}

// GetDefaultSchema returns a copy of the official Azure schema for a given service slug.
// If the slug is embedded in the binary, it returns immediately (0ms).
// If not embedded, it fetches dynamically on-demand from Azure's public API and caches it.
func GetDefaultSchema(slug string) map[string]interface{} {
	if slug == "virtual-machines" || slug == "azure-sql-database" {
		return nil // Handled directly with compact, zero-latency dedicated builders
	}

	schemasMu.RLock()
	if s, ok := defaultSchemas[slug]; ok {
		schemasMu.RUnlock()
		return cloneMap(s)
	}
	schemasMu.RUnlock()

	// Fetch dynamically from Azure pricing API
	onlineSchema := FetchSchemaOnline(slug)
	if onlineSchema != nil {
		schemasMu.Lock()
		defaultSchemas[slug] = onlineSchema
		schemasMu.Unlock()
		return cloneMap(onlineSchema)
	}

	return nil
}

// FetchSchemaOnline queries Azure's public calculator pricing API for any unknown service slug
func FetchSchemaOnline(slug string) map[string]interface{} {
	urls := []string{
		fmt.Sprintf("https://azure.microsoft.com/api/v3/pricing/%s/calculator/?culture=en-us", slug),
		fmt.Sprintf("https://azure.microsoft.com/api/v2/pricing/%s/calculator/?culture=en-us", slug),
	}

	client := &http.Client{Timeout: 6 * time.Second}
	for _, u := range urls {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var res struct {
				Schema map[string]interface{} `json:"schema"`
			}
			err := json.NewDecoder(resp.Body).Decode(&res)
			resp.Body.Close()
			if err == nil && res.Schema != nil {
				return res.Schema
			}
		} else {
			resp.Body.Close()
		}
	}
	return nil
}

// ListEmbeddedServices returns all service slugs currently pre-compiled in the binary
func ListEmbeddedServices() []string {
	schemasMu.RLock()
	defer schemasMu.RUnlock()

	var list []string
	for k := range defaultSchemas {
		list = append(list, k)
	}
	return list
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{}, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
