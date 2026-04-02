package gravatar

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

const baseURL = "https://gravatar.com/"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9_.%+\-]+@[a-zA-Z0-9\-]+\.[a-zA-Z0-9\-.]+$`)

// Client handles HTTP requests to Gravatar.
type Client struct {
	HTTP    *http.Client
	BaseURL string
}

// NewClient creates a Gravatar HTTP client.
func NewClient() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		BaseURL: baseURL,
	}
}

// Profile represents JSON API data from Gravatar.
type Profile struct {
	Hash              string            `json:"hash"`
	ThumbnailURL      string            `json:"thumbnailUrl"`
	PreferredUsername  string            `json:"preferredUsername"`
	DisplayName       string            `json:"displayName"`
	Pronunciation     string            `json:"pronunciation"`
	AboutMe           string            `json:"aboutMe"`
	CurrentLocation   string            `json:"currentLocation"`
	JobTitle          string            `json:"jobTitle"`
	Company           string            `json:"company"`
	LastProfileEdit   string            `json:"lastProfileEdit"`
	Pronouns          string            `json:"pronouns"`
	Name              map[string]string `json:"name"`
	Emails            []EmailEntry      `json:"emails"`
	ContactInfo       json.RawMessage   `json:"contactInfo"`
	PhoneNumbers      json.RawMessage   `json:"phoneNumbers"`
}

// EmailEntry represents an email in the Gravatar JSON response.
type EmailEntry struct {
	Value string `json:"value"`
}

// ScrapedData contains data scraped from the Gravatar HTML profile page.
type ScrapedData struct {
	Accounts  []VerifiedAccount
	Photos    []string
	Payments  []Payment
	Interests []string
	Links     []Link
}

// VerifiedAccount represents a verified social account.
type VerifiedAccount struct {
	Network string
	URL     string
}

// Payment represents a payment method on the profile.
type Payment struct {
	Title string
	Asset string
}

// Link represents a link on the profile.
type Link struct {
	Name        string
	URL         string
	Description string
}

// FullProfile combines JSON and scraped data.
type FullProfile struct {
	Hash             string
	ProfileURL       string
	Avatar           string
	LastEdit         string
	Location         string
	PreferredUsername string
	DisplayName      string
	Pronunciation    string
	Name             map[string]string
	Pronouns         string
	AboutMe          string
	JobTitle         string
	Company          string
	Emails           []string
	ContactInfo      json.RawMessage
	PhoneNumbers     json.RawMessage
	Accounts         []VerifiedAccount
	Payments         []Payment
	Photos           []string
	Interests        []string
	Links            []Link
}

type jsonResponse struct {
	Entry []Profile `json:"entry"`
}

// FetchJSON fetches the Gravatar JSON profile for a given ID (hash or username).
func (c *Client) FetchJSON(id string) (*Profile, error) {
	url := c.BaseURL + id + ".json"
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("gravatar profile not found (404)")
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("too many requests, try again later (429)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB limit
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var jr jsonResponse
	if err := json.Unmarshal(body, &jr); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	if len(jr.Entry) == 0 {
		return nil, fmt.Errorf("no entries in response")
	}

	return &jr.Entry[0], nil
}

// ScrapeProfile scrapes the Gravatar HTML profile page.
func (c *Client) ScrapeProfile(id string) (*ScrapedData, error) {
	url := c.BaseURL + id
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	return ScrapeHTML(resp.Body)
}

// ScrapeHTML parses Gravatar profile HTML and extracts structured data.
func ScrapeHTML(r io.Reader) (*ScrapedData, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	data := &ScrapedData{}
	data.Accounts = findAccounts(doc)
	data.Photos = findPhotos(doc)
	data.Payments = findPayments(doc)
	data.Interests = findInterests(doc)
	data.Links = findLinks(doc)

	return data, nil
}

func findAccounts(doc *goquery.Document) []VerifiedAccount {
	var accounts []VerifiedAccount
	doc.Find(".is-verified-accounts .card-item__info").Each(func(_ int, s *goquery.Selection) {
		network := strings.TrimSpace(s.Find(".card-item__label-text").Text())
		s.Find("a").Each(func(_ int, a *goquery.Selection) {
			classes, _ := a.Attr("class")
			if strings.Contains(classes, "card-item__checkmark-icon") {
				return
			}
			href, exists := a.Attr("href")
			if exists && href != "" {
				accounts = append(accounts, VerifiedAccount{Network: network, URL: href})
			}
		})
	})
	return accounts
}

func findPhotos(doc *goquery.Document) []string {
	var photos []string
	doc.Find(".g-profile__photo-gallery img").Each(func(_ int, s *goquery.Selection) {
		dataURL, exists := s.Attr("data-url")
		if exists && dataURL != "" {
			photos = append(photos, dataURL+"?size=666")
		}
	})
	return photos
}

func findPayments(doc *goquery.Document) []Payment {
	var payments []Payment
	doc.Find(".payments-drawer .card-item").Each(func(_ int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".card-item__label-text").Text())
		a := s.Find("a")
		if href, exists := a.Attr("href"); exists && href != "" {
			payments = append(payments, Payment{Title: title, Asset: href})
		} else {
			span := s.Find(".card-item__info span").Not(".card-item__label-text")
			asset := strings.TrimSpace(span.First().Text())
			if asset != "" {
				payments = append(payments, Payment{Title: title, Asset: asset})
			}
		}
	})
	return payments
}

func findInterests(doc *goquery.Document) []string {
	var interests []string
	doc.Find(".g-profile__interests-list li a, .g-profile__interests-list li span").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			interests = append(interests, text)
		}
	})
	return interests
}

func findLinks(doc *goquery.Document) []Link {
	var links []Link
	doc.Find(".g-profile__links .card-item__info").Each(func(_ int, s *goquery.Selection) {
		a := s.Find("a").First()
		if a.Length() == 0 {
			return
		}
		name := strings.TrimSpace(a.Text())
		runes := []rune(name)
		if len(runes) >= 2 {
			name = string(runes[:len(runes)-2])
		}
		href, _ := a.Attr("href")
		desc := strings.TrimSpace(s.Find("p").Text())
		links = append(links, Link{Name: name, URL: href, Description: desc})
	})
	return links
}

// AggregateProfile fetches and combines JSON + scraped data.
func (c *Client) AggregateProfile(id string) (*FullProfile, error) {
	profile, err := c.FetchJSON(id)
	if err != nil {
		return nil, err
	}

	scraped, err := c.ScrapeProfile(id)
	if err != nil {
		// Non-fatal: continue with JSON data only
		scraped = &ScrapedData{}
	}

	var emails []string
	for _, e := range profile.Emails {
		emails = append(emails, e.Value)
	}

	fp := &FullProfile{
		Hash:             profile.Hash,
		ProfileURL:       baseURL + profile.Hash,
		Avatar:           profile.ThumbnailURL + "?size=666",
		LastEdit:         profile.LastProfileEdit,
		Location:         profile.CurrentLocation,
		PreferredUsername: profile.PreferredUsername,
		DisplayName:      profile.DisplayName,
		Pronunciation:    profile.Pronunciation,
		Name:             profile.Name,
		Pronouns:         profile.Pronouns,
		AboutMe:          profile.AboutMe,
		JobTitle:         profile.JobTitle,
		Company:          profile.Company,
		Emails:           emails,
		ContactInfo:      profile.ContactInfo,
		PhoneNumbers:     profile.PhoneNumbers,
		Accounts:         scraped.Accounts,
		Payments:         scraped.Payments,
		Photos:           scraped.Photos,
		Interests:        scraped.Interests,
		Links:            scraped.Links,
	}
	return fp, nil
}

// DisplayProfile prints a FullProfile as a formatted table.
func DisplayProfile(fp *FullProfile) {
	cyan := color.New(color.FgCyan, color.Bold)

	cyan.Printf("\n  %s\n\n", fp.PreferredUsername)

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	printRow := func(key, value string) {
		if value != "" {
			fmt.Fprintf(w, "  %s\t%s\n", color.CyanString(key), value)
		}
	}

	printRow("Hash", fp.Hash)
	printRow("Profile URL", fp.ProfileURL)
	printRow("Avatar", fp.Avatar)
	printRow("Last edit", fp.LastEdit)
	printRow("Location", fp.Location)
	printRow("Preferred username", fp.PreferredUsername)
	printRow("Display name", fp.DisplayName)
	printRow("Pronunciation", fp.Pronunciation)
	if fp.Name != nil {
		for k, v := range fp.Name {
			printRow("Name ("+k+")", v)
		}
	}
	printRow("Pronouns", fp.Pronouns)
	printRow("About me", fp.AboutMe)
	printRow("Job Title", fp.JobTitle)
	printRow("Company", fp.Company)

	if len(fp.Emails) > 0 {
		printRow("Emails", strings.Join(fp.Emails, ", "))
	}

	if len(fp.Accounts) > 0 {
		var parts []string
		for _, a := range fp.Accounts {
			parts = append(parts, fmt.Sprintf("%s: %s", a.Network, a.URL))
		}
		printRow("Verified accounts", strings.Join(parts, "\n\t"))
	}

	if len(fp.Payments) > 0 {
		var parts []string
		for _, p := range fp.Payments {
			parts = append(parts, fmt.Sprintf("%s: %s", p.Title, p.Asset))
		}
		printRow("Payments", strings.Join(parts, "\n\t"))
	}

	if len(fp.Photos) > 0 {
		printRow("Photos", strings.Join(fp.Photos, "\n\t"))
	}

	if len(fp.Interests) > 0 {
		printRow("Interests", strings.Join(fp.Interests, ", "))
	}

	if len(fp.Links) > 0 {
		var parts []string
		for _, l := range fp.Links {
			parts = append(parts, fmt.Sprintf("%s: %s", l.Name, l.URL))
		}
		printRow("Links", strings.Join(parts, "\n\t"))
	}

	w.Flush()
	fmt.Println()
}

// ValidateEmail checks if a string is a valid email format.
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// HashEmail returns the MD5 hex digest of a trimmed, lowercased email.
func HashEmail(email string) string {
	h := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("%x", h)
}

// HashEmailSHA256 returns the SHA256 hex digest of a trimmed, lowercased email.
func HashEmailSHA256(email string) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("%x", h)
}
