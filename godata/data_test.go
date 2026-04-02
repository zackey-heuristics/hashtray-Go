package godata

import "testing"

func TestLoadDomains(t *testing.T) {
	tests := []struct {
		listType string
		wantLen  int
	}{
		{"common", 455},
		{"", 455}, // default
		{"long", 5334},
		{"full", 118062},
	}

	for _, tt := range tests {
		t.Run(tt.listType, func(t *testing.T) {
			domains, err := LoadDomains(tt.listType)
			if err != nil {
				t.Fatalf("LoadDomains(%q) error: %v", tt.listType, err)
			}
			if len(domains) != tt.wantLen {
				t.Errorf("LoadDomains(%q) returned %d domains, want %d", tt.listType, len(domains), tt.wantLen)
			}
		})
	}
}

func TestLoadDomainsInvalid(t *testing.T) {
	_, err := LoadDomains("nonexistent")
	if err == nil {
		t.Error("LoadDomains(\"nonexistent\") should return error")
	}
}
