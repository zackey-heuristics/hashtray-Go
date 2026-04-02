package enumerator

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/publicsuffix"

	"github.com/zackey-heuristics/hashtray-Go/godata"
	"github.com/zackey-heuristics/hashtray-Go/internal/elements"
	"github.com/zackey-heuristics/hashtray-Go/internal/gravatar"
	"github.com/zackey-heuristics/hashtray-Go/internal/permutator"
)

var (
	md5Regex    = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	sha256Regex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	emailRegex  = regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`)
)

// Options configures the enumerator.
type Options struct {
	DomainList    string   // "common", "long", "full"
	Elements      []string // user-provided elements
	CustomDomains []string // user-provided domains
	Crazy         bool
}

// Enumerator orchestrates email enumeration from a Gravatar account.
type Enumerator struct {
	account     string
	opts        Options
	client      *gravatar.Client
	accountHash string
	hashType    string
	profile     *gravatar.FullProfile
	chunks      []string
	domains     []string
	publicEmails []string
}

// New creates a new Enumerator.
func New(account string, opts Options) *Enumerator {
	return &Enumerator{
		account: account,
		opts:    opts,
		client:  gravatar.NewClient(),
	}
}

// DetectHashType returns "MD5", "SHA256", or "" for a given string.
func DetectHashType(s string) string {
	if md5Regex.MatchString(s) {
		return "MD5"
	}
	if sha256Regex.MatchString(s) {
		return "SHA256"
	}
	return ""
}

// Run executes the enumeration.
func (e *Enumerator) Run() error {
	e.hashType = DetectHashType(e.account)

	if err := e.resolveAccount(); err != nil {
		return err
	}

	// Load domains
	domains, err := godata.LoadDomains(e.opts.DomainList)
	if err != nil {
		return fmt.Errorf("loading domains: %w", err)
	}
	e.domains = domains

	// Add custom domains at front
	if len(e.opts.CustomDomains) > 0 {
		e.domains = append(e.opts.CustomDomains, e.domains...)
	}

	if e.profile != nil {
		e.extractPublicEmails()
		e.extractElements()
	}

	// Add user-provided elements
	if len(e.opts.Elements) > 0 {
		for _, el := range e.opts.Elements {
			if !contains(e.chunks, el) {
				e.chunks = append(e.chunks, el)
			}
		}
	}

	if len(e.chunks) == 0 {
		return fmt.Errorf("no elements to permute")
	}

	// Create permutator
	perm := permutator.New(e.chunks, e.domains, e.opts.Crazy)
	count := perm.CombinationCount()

	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Printf("Elements to permute: ")
	yellow.Println(strings.Join(e.chunks, ", "))
	fmt.Printf("Number of email domains: %d\n", len(e.domains))
	fmt.Printf("Number of possible combinations: %d\n\n", count)

	// Hash function
	hasher := e.getHasher()

	// Enumerate
	bar := progressbar.NewOptions(count,
		progressbar.OptionSetDescription("Comparing email hashes"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var found string
	for email := range perm.GenerateCtx(ctx) {
		hashed := hasher(email)
		bar.Add(1) //nolint:errcheck
		if hashed == e.accountHash {
			found = email
			cancel()
			break
		}
	}
	fmt.Println()

	// Results
	cyan.Println("\nRESULTS:")

	if len(e.publicEmails) > 0 {
		label := "Email"
		if len(e.publicEmails) > 1 {
			label = "Emails"
		}
		fmt.Printf("\n%s found in the public profile: %s\n\n",
			color.CyanString(label),
			color.HiWhiteString(strings.Join(e.publicEmails, ", ")))

		for _, pub := range e.publicEmails {
			if e.accountHash == hasher(pub) {
				color.Green("%s matches the account hash. It's used as the primary Gravatar email for the account", pub)
			} else {
				color.Yellow("%s does not match the account hash. The Gravatar account email is not this email, there is at least another one to find.", pub)
			}
		}
	}

	if found != "" {
		fmt.Printf("\n%s %s\n\n",
			color.HiWhiteString("An email has been found with the email hashes enumeration:"),
			color.GreenString(found))

		if e.profile != nil {
			fmt.Print("Do you want to see the gravatar profile if available? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "y" || answer == "yes" {
				gravatar.DisplayProfile(e.profile)
			}
		}
	} else if len(e.publicEmails) == 0 {
		color.Red("\nNo email found matching the account hash: %s", e.accountHash)
	}

	fmt.Println()
	return nil
}

func (e *Enumerator) resolveAccount() error {
	red := color.New(color.FgRed)

	if e.hashType != "" {
		// Account is a hash
		e.accountHash = e.account
		profile, err := e.client.AggregateProfile(e.account)
		if err != nil {
			if len(e.opts.Elements) == 0 {
				red.Printf("A matching Gravatar account for %s could not be retrieved.\n\n", e.account)
				fmt.Println("To continue, you'll need the account's hash and some known elements about the target.")
				fmt.Println("Use: hashtray account <HASH> -e element1 element2 ...")
				return fmt.Errorf("no gravatar profile found")
			}
			red.Printf("No Gravatar account found for the provided hash: %s.\n", e.account)
			color.Yellow("Continuing with the provided elements to search for possible email addresses.\n")
			return nil
		}
		e.profile = profile
	} else {
		// Account is a username
		profile, err := e.client.AggregateProfile(e.account)
		if err != nil {
			red.Printf("A matching Gravatar account for %s could not be retrieved.\n\n", e.account)
			fmt.Println("To continue, you'll need the account's hash and some known elements about the target.")
			fmt.Println("Use: hashtray account <HASH> -e element1 element2 ...")
			return fmt.Errorf("no gravatar profile found")
		}
		e.profile = profile
		e.accountHash = profile.Hash
		e.hashType = DetectHashType(e.accountHash)
		if e.hashType == "" {
			e.hashType = "MD5" // default
		}
	}
	return nil
}

func (e *Enumerator) extractPublicEmails() {
	if e.profile == nil {
		return
	}
	e.publicEmails = append(e.publicEmails, e.profile.Emails...)

	// Find emails in AboutMe
	if e.profile.AboutMe != "" {
		found := emailRegex.FindAllString(e.profile.AboutMe, -1)
		e.publicEmails = append(e.publicEmails, found...)
	}
}

func (e *Enumerator) extractElements() {
	if e.profile == nil {
		return
	}

	// Convert FullProfile to elements.Profile
	var accounts []elements.Account
	for _, a := range e.profile.Accounts {
		accounts = append(accounts, elements.Account{Network: a.Network, URL: a.URL})
	}
	var links []elements.Link
	for _, l := range e.profile.Links {
		links = append(links, elements.Link{Name: l.Name, URL: l.URL, Description: l.Description})
	}

	elemProfile := &elements.Profile{
		PreferredUsername: e.profile.PreferredUsername,
		DisplayName:      e.profile.DisplayName,
		VerifiedAccounts: accounts,
		Links:            links,
		Emails:           e.profile.Emails,
		AboutMe:          e.profile.AboutMe,
	}

	elems, extraDomains := elements.Extract(elemProfile)
	e.chunks = elems

	// Add link domains at front (extract eTLD+1 like Python's tldextract)
	if e.profile.Links != nil {
		for _, link := range e.profile.Links {
			if link.URL != "" {
				if parsed, err := url.Parse(link.URL); err == nil && parsed.Host != "" {
					if eTLD1, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname()); err == nil {
						if !contains(e.domains, eTLD1) {
							e.domains = append([]string{eTLD1}, e.domains...)
						}
					}
				}
			}
		}
	}

	// Add element-extracted domains
	for _, d := range extraDomains {
		if !contains(e.domains, d) {
			e.domains = append([]string{d}, e.domains...)
		}
	}
}

func (e *Enumerator) getHasher() func(string) string {
	if e.hashType == "SHA256" {
		return gravatar.HashEmailSHA256
	}
	return gravatar.HashEmail
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
