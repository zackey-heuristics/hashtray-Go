package gravatar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"user@domain", false},
		{"@domain.com", false},
		{"user@", false},
		{"", false},
		{"noat", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

func TestHashEmail(t *testing.T) {
	// MD5 of "user@example.com"
	got := HashEmail("User@Example.COM")
	want := "b58996c504c5638798eb6b511e6f49af"
	if got != want {
		t.Errorf("HashEmail = %s, want %s", got, want)
	}
}

func TestHashEmailSHA256(t *testing.T) {
	got := HashEmailSHA256("User@Example.COM")
	// SHA256 of "user@example.com"
	if len(got) != 64 {
		t.Errorf("HashEmailSHA256 should produce 64 char hex, got %d", len(got))
	}
}

func TestFetchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/testuser.json" {
			resp := map[string]interface{}{
				"entry": []map[string]interface{}{
					{
						"hash":              "abc123",
						"preferredUsername":  "testuser",
						"displayName":       "Test User",
						"thumbnailUrl":      "https://gravatar.com/avatar/abc123",
						"lastProfileEdit":   "2024-01-01",
						"currentLocation":   "Paris",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := &Client{HTTP: server.Client(), BaseURL: server.URL + "/"}
	profile, err := client.FetchJSON("testuser")
	if err != nil {
		t.Fatalf("FetchJSON error: %v", err)
	}
	if profile.Hash != "abc123" {
		t.Errorf("Hash = %s, want abc123", profile.Hash)
	}
	if profile.PreferredUsername != "testuser" {
		t.Errorf("PreferredUsername = %s, want testuser", profile.PreferredUsername)
	}
}

func TestFetchJSON404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	client := &Client{HTTP: server.Client(), BaseURL: server.URL + "/"}
	_, err := client.FetchJSON("nonexistent")
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestScrapeHTML(t *testing.T) {
	html := `<html><body>
		<div class="is-verified-accounts">
			<div class="card-item__info">
				<span class="card-item__label-text">Twitter</span>
				<a href="https://twitter.com/testuser">testuser</a>
			</div>
		</div>
		<div class="g-profile__photo-gallery">
			<img data-url="https://gravatar.com/avatar/abc123" />
		</div>
		<div class="g-profile__interests-list">
			<li><a>Photography</a></li>
			<li><span>Coding</span></li>
		</div>
		<div class="g-profile__links">
			<div class="card-item__info">
				<a href="https://example.com">My Blog ↗</a>
				<p>Personal website</p>
			</div>
		</div>
	</body></html>`

	data, err := ScrapeHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ScrapeHTML error: %v", err)
	}

	if len(data.Accounts) != 1 || data.Accounts[0].Network != "Twitter" {
		t.Errorf("Accounts = %+v, want 1 Twitter account", data.Accounts)
	}

	if len(data.Photos) != 1 {
		t.Errorf("Photos = %+v, want 1 photo", data.Photos)
	}

	if len(data.Interests) != 2 {
		t.Errorf("Interests = %+v, want 2 interests", data.Interests)
	}

	if len(data.Links) != 1 || data.Links[0].Description != "Personal website" {
		t.Errorf("Links = %+v, want 1 link with description", data.Links)
	}
}
