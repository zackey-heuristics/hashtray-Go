package elements

import "testing"

func TestExtractBasic(t *testing.T) {
	profile := &Profile{
		PreferredUsername: "jdoe",
		DisplayName:      "John Doe",
		VerifiedAccounts: []Account{
			{Network: "Twitter", URL: "https://twitter.com/john_doe"},
			{Network: "LinkedIn", URL: "https://linkedin.com/in/johndoe"},
		},
	}

	elems, _ := Extract(profile)

	// "jdoe" gets deduped as combination of "j" + "doe", matching Python behavior
	want := map[string]bool{
		"john": true,
		"doe":  true,
		"j":    true,
		"d":    true,
	}

	for _, e := range elems {
		delete(want, e)
	}
	for missing := range want {
		t.Errorf("missing expected element: %s", missing)
	}
}

func TestExtractBluesky(t *testing.T) {
	profile := &Profile{
		DisplayName: "Test",
		VerifiedAccounts: []Account{
			{Network: "Bluesky", URL: "https://bsky.app/profile/alice.bsky.social"},
		},
	}
	elems, _ := Extract(profile)
	found := false
	for _, e := range elems {
		if e == "alice" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'alice' from Bluesky handle")
	}
}

func TestExtractWordPressDomain(t *testing.T) {
	profile := &Profile{
		DisplayName: "Test",
		VerifiedAccounts: []Account{
			{Network: "WordPress", URL: "https://myblog.example.com"},
		},
	}
	elems, _ := Extract(profile)

	// WordPress with subdomain should add the subdomain as an element
	foundSubdomain := false
	for _, e := range elems {
		if e == "myblog" {
			foundSubdomain = true
		}
	}
	if !foundSubdomain {
		t.Errorf("expected 'myblog' element from WordPress subdomain, got elements: %v", elems)
	}
}

func TestIsCombination(t *testing.T) {
	chunks := []string{"john", "doe", "johndoe"}
	if !isCombination("johndoe", chunks) {
		t.Error("johndoe should be a combination of john + doe")
	}
	if isCombination("john", chunks) {
		t.Error("john should not be a combination")
	}
}

func TestDedupChunks(t *testing.T) {
	profile := &Profile{
		PreferredUsername: "johndoe",
		DisplayName:      "John Doe",
	}
	elems, _ := Extract(profile)
	for _, e := range elems {
		if e == "johndoe" {
			t.Error("johndoe should be deduped as combination of john+doe")
		}
	}
}

func TestExtractEmpty(t *testing.T) {
	profile := &Profile{}
	elems, domains := Extract(profile)
	if len(elems) != 0 {
		t.Errorf("expected 0 elements, got %d", len(elems))
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}
