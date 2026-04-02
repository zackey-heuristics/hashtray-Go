package godata

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed email_services.json email_services_long.json email_services_full.json
var dataFS embed.FS

// LoadDomains loads the email domain list for the given list type.
// Valid types: "common" (default), "long", "full".
func LoadDomains(listType string) ([]string, error) {
	filename, err := domainFilename(listType)
	if err != nil {
		return nil, err
	}

	data, err := dataFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading embedded file %s: %w", filename, err)
	}

	var domains []string
	if err := json.Unmarshal(data, &domains); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	return domains, nil
}

func domainFilename(listType string) (string, error) {
	switch listType {
	case "", "common":
		return "email_services.json", nil
	case "long":
		return "email_services_long.json", nil
	case "full":
		return "email_services_full.json", nil
	default:
		return "", fmt.Errorf("unknown domain list type: %q (valid: common, long, full)", listType)
	}
}
