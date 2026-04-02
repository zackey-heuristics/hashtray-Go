package elements

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/mozillazg/go-unidecode"
	"golang.org/x/net/publicsuffix"
)

// Account represents a verified social account on a Gravatar profile.
type Account struct {
	Network string
	URL     string
}

// Link represents a link on a Gravatar profile.
type Link struct {
	Name        string
	URL         string
	Description string
}

// Profile holds the Gravatar profile data needed for element extraction.
type Profile struct {
	PreferredUsername string
	DisplayName      string
	VerifiedAccounts []Account
	Links            []Link
	Emails           []string
	AboutMe          string
	Hash             string
	ProfileURL       string
	Avatar           string
	LastEdit         string
	Location         string
	Pronunciation    string
	Name             map[string]string
	Pronouns         string
	JobTitle         string
	Company          string
	ContactInfo      interface{}
	PhoneNumbers     interface{}
	Payments         interface{}
	Photos           []string
	Interests        []string
}

var namePattern = regexp.MustCompile(`[-_ ./]`)

// Extract returns elements and domains from a Gravatar profile.
func Extract(profile *Profile) ([]string, []string) {
	e := &extractor{
		profile: profile,
	}
	e.addPreferredUsername()
	e.addDisplayName()
	e.addAccounts()
	e.formatElements()
	e.elements = e.dedupChunks()
	return e.elements, e.domains
}

type extractor struct {
	profile  *Profile
	elements []string
	domains  []string
}

func (e *extractor) addPreferredUsername() {
	if e.profile.PreferredUsername != "" {
		e.elements = append(e.elements, e.profile.PreferredUsername)
	}
}

func (e *extractor) addDisplayName() {
	if e.profile.DisplayName == "" {
		return
	}
	decoded := unidecode.Unidecode(e.profile.DisplayName)
	names := namePattern.Split(decoded, -1)
	for _, name := range names {
		name = strings.ReplaceAll(name, "\"", "")
		name = strings.ReplaceAll(name, "'", "")
		if name != "" {
			e.elements = append(e.elements, name)
			if len(name) > 1 {
				e.elements = append(e.elements, name[:1])
			}
		}
	}
}

func (e *extractor) addAccounts() {
	if e.profile.VerifiedAccounts == nil {
		return
	}
	for _, acct := range e.profile.VerifiedAccounts {
		accountURL := strings.TrimRight(acct.URL, "/")
		e.processAccount(acct.Network, accountURL)
	}
}

func (e *extractor) processAccount(network, accountURL string) {
	chunk := lastURLChunk(accountURL)

	switch network {
	case "Mastodon", "Fediverse", "TikTok":
		e.elements = append(e.elements, strings.ReplaceAll(chunk, "@", ""))

	case "LinkedIn":
		if strings.Contains(accountURL, "linkedin.com/in/") {
			e.elements = append(e.elements, chunk)
		}

	case "YouTube":
		e.elements = append(e.elements, strings.TrimLeft(chunk, "@"))

	case "Tumblr":
		sub, _, _ := extractTLD(accountURL)
		if sub != "" {
			e.elements = append(e.elements, sub)
		} else {
			e.elements = append(e.elements, lastURLChunk(accountURL))
		}

	case "WordPress":
		sub, domain, suffix := extractTLD(accountURL)
		if sub != "" {
			e.elements = append(e.elements, sub)
		} else if domain != "" {
			e.elements = append(e.elements, domain)
			if suffix != "" {
				e.domains = append(e.domains, domain+"."+suffix)
			}
		}

	case "Bluesky":
		handle := lastURLChunk(accountURL)
		if strings.HasSuffix(handle, ".bsky.social") {
			e.elements = append(e.elements, strings.TrimSuffix(handle, ".bsky.social"))
		} else {
			parts := strings.SplitN(handle, ".", 2)
			e.elements = append(e.elements, parts[0])
			if len(parts) == 2 {
				e.domains = append(e.domains, handle)
			}
		}

	case "Facebook", "Instagram":
		if !strings.Contains(accountURL, "profile.php") {
			for _, part := range strings.Split(chunk, ".") {
				if part != "" {
					e.elements = append(e.elements, part)
				}
			}
		}

	case "Stack Overflow":
		for _, part := range strings.Split(chunk, "-") {
			if part != "" {
				e.elements = append(e.elements, part)
			}
		}

	case "Flickr":
		if !strings.Contains(accountURL, "/people/") {
			for _, part := range strings.Split(chunk, "-") {
				if part != "" {
					e.elements = append(e.elements, part)
				}
			}
		}

	case "Twitter", "X":
		for _, part := range strings.Split(chunk, "_") {
			if part != "" {
				e.elements = append(e.elements, part)
			}
		}

	case "TripIt":
		if strings.Contains(accountURL, "/people/") {
			for _, part := range strings.Split(chunk, ".") {
				if part != "" {
					e.elements = append(e.elements, part)
				}
			}
		}

	case "Goodreads":
		parts := strings.Split(chunk, "-")
		if len(parts) > 1 {
			for _, part := range parts[1:] {
				if part != "" {
					e.elements = append(e.elements, part)
				}
			}
		}

	case "Foursquare", "Yahoo", "Google+", "Vimeo":
		// skip

	default:
		if chunk != "" {
			e.elements = append(e.elements, chunk)
		}
	}
}

func (e *extractor) formatElements() {
	seen := make(map[string]bool)
	var unique []string
	for _, el := range e.elements {
		lower := strings.ToLower(unidecode.Unidecode(el))
		if !seen[lower] {
			seen[lower] = true
			unique = append(unique, lower)
		}
	}
	e.elements = unique
}

func (e *extractor) dedupChunks() []string {
	var result []string
	for _, s := range e.elements {
		chunks := make([]string, len(e.elements))
		copy(chunks, e.elements)
		if !isCombination(s, chunks) {
			result = append(result, s)
		}
	}
	return result
}

// isCombination checks if s can be formed by concatenating other strings in chunks.
func isCombination(s string, chunks []string) bool {
	// Remove s itself from chunks
	filtered := removeFirst(chunks, s)

	for i := 1; i < len(s); i++ {
		left := s[:i]
		right := s[i:]
		if contains(filtered, left) && contains(filtered, right) {
			return true
		}
		if contains(filtered, left) && isCombination(right, filtered) {
			return true
		}
		if contains(filtered, right) && isCombination(left, filtered) {
			return true
		}
	}
	return false
}

func removeFirst(items []string, target string) []string {
	result := make([]string, 0, len(items))
	removed := false
	for _, item := range items {
		if !removed && item == target {
			removed = true
			continue
		}
		result = append(result, item)
	}
	return result
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func lastURLChunk(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	return parts[len(parts)-1]
}

// extractTLD extracts subdomain, domain, and suffix from a URL.
func extractTLD(rawURL string) (subdomain, domain, suffix string) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", "", ""
	}
	host := parsed.Hostname()

	eTLD, icann := publicsuffix.PublicSuffix(host)
	if !icann && !strings.Contains(eTLD, ".") {
		// Not a known public suffix; treat last part as suffix
		eTLD = ""
	}

	if eTLD == "" {
		parts := strings.SplitN(host, ".", 2)
		if len(parts) == 2 {
			return "", parts[0], parts[1]
		}
		return "", host, ""
	}

	// Remove eTLD from host
	withoutSuffix := strings.TrimSuffix(host, "."+eTLD)
	parts := strings.Split(withoutSuffix, ".")
	if len(parts) == 1 {
		return "", parts[0], eTLD
	}
	domain = parts[len(parts)-1]
	subdomain = strings.Join(parts[:len(parts)-1], ".")
	suffix = eTLD
	return
}
