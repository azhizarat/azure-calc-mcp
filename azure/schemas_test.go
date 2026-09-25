package azure

import (
	"testing"
)

func TestGetDefaultSchemaEmbedded(t *testing.T) {
	services := []string{
		"kubernetes-service",
		"postgresql",
		"storage",
		"key-vault",
		"monitor",
		"cosmos-db",
		"cognitive-services",
		"container-apps",
		"application-gateway",
		"azure-firewall",
	}

	for _, s := range services {
		sc := GetDefaultSchema(s)
		if sc == nil {
			t.Errorf("expected embedded schema for %q, got nil", s)
			continue
		}
		if len(sc) == 0 {
			t.Errorf("expected non-empty schema for %q", s)
		}
	}
}

func TestSchemaCloneMapIndependence(t *testing.T) {
	sc1 := GetDefaultSchema("storage")
	if sc1 == nil {
		t.Fatal("expected storage schema")
	}

	sc1["customKey"] = "testValue"

	sc2 := GetDefaultSchema("storage")
	if _, exists := sc2["customKey"]; exists {
		t.Errorf("mutation of sc1 affected sc2! cloneMap failed to isolate instances")
	}
}

func TestListEmbeddedServices(t *testing.T) {
	list := ListEmbeddedServices()
	if len(list) < 15 {
		t.Errorf("expected at least 15 embedded services, got %d", len(list))
	}
}
