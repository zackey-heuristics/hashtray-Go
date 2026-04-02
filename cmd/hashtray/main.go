package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zackey-heuristics/hashtray-Go/internal/enumerator"
	"github.com/zackey-heuristics/hashtray-Go/internal/gravatar"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "hashtray-go",
		Short:   "Gravatar Account and Email Finder",
		Long:    banner(),
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			printUsage()
			cmd.Help() //nolint:errcheck
		},
	}

	// email subcommand
	emailCmd := &cobra.Command{
		Use:   "email [address]",
		Short: "Find a Gravatar account from an email address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			email := strings.TrimSpace(args[0])
			if !gravatar.ValidateEmail(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}

			fmt.Println("Retrieving and scraping profile...")

			client := gravatar.NewClient()
			hash := gravatar.HashEmail(email)
			profile, err := client.AggregateProfile(hash)
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}

			// Check if the lookup hash differs from the profile hash
			if hash != profile.Hash {
				color.Yellow("Note: The email hash (%s) differs from the profile hash (%s).", hash, profile.Hash)
				color.Yellow("This email is likely a secondary email on the Gravatar account.")
				fmt.Println("Use 'hashtray-go account' to find the primary email.")
				fmt.Println()
			}

			gravatar.DisplayProfile(profile)
			return nil
		},
	}

	// account subcommand
	var domainList string
	var elementsFlag []string
	var domainsFlag []string
	var crazy bool

	accountCmd := &cobra.Command{
		Use:   "account [username or hash]",
		Short: "Find an email address from a Gravatar username or hash (MD5/SHA256)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := enumerator.Options{
				DomainList:    domainList,
				Elements:      elementsFlag,
				CustomDomains: domainsFlag,
				Crazy:         crazy,
			}
			e := enumerator.New(args[0], opts)
			return e.Run()
		},
	}

	accountCmd.Flags().StringVarP(&domainList, "domain_list", "l", "common",
		"Domain list to use: common, long, full")
	accountCmd.Flags().StringSliceVarP(&elementsFlag, "elements", "e", nil,
		"Custom elements/strings for email generation")
	accountCmd.Flags().StringSliceVarP(&domainsFlag, "domains", "d", nil,
		"Custom email domains")
	accountCmd.Flags().BoolVarP(&crazy, "crazy", "c", false,
		"Try EVERY SINGLE combination (with any special char at any place)")

	rootCmd.AddCommand(emailCmd, accountCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func banner() string {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	return fmt.Sprintf(`
%s
%s
%s
%s
%s
%s  v%s

Gravatar Account and Email Finder
`, cyan("⠀⠀⣠⣴⣶⠿⠿⣷⣶⣄⠀⠀⠀"),
		cyan("⠀⢾⠟⠉⠀⠀⠀⠀⠈⠁⠀⠀⠀"),
		cyan("⠀⠀⠀⠀⠀⠀⣶⣶⣶⣶⣶⣶⠂"),
		cyan("⣀⣀⠀⠀⠀⠀⠋⠃⠀⠀⢸⣿"),
		cyan("⠘⢿⣦⣀⠀⠀⠀⠀⢀⣴⣿⠋"),
		cyan("⠀⠀⠙⠻⠿⣶⣶⡿⠿⠋⠁"),
		version)
}

func printUsage() {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgHiRed)

	cyan.Println("Gravatar Account and Email Finder")
	fmt.Println("1) Find a Gravatar account from an email address.")
	fmt.Println("2) Find an email address from a Gravatar account or hash.")
	fmt.Println()

	fmt.Print("▶ ")
	cyan.Print("Find a gravatar account from an email: ")
	fmt.Println()
	fmt.Printf("  Usage: ")
	yellow.Print("hashtray-go ")
	red.Println("email user@example.com")
	fmt.Println()

	fmt.Print("▶ ")
	cyan.Print("Find a gravatar email from a username or hash: ")
	fmt.Println()
	fmt.Printf("  Usage: ")
	yellow.Print("hashtray-go ")
	red.Println("account username")
	fmt.Printf("         ")
	yellow.Print("hashtray-go ")
	red.Println("account cc8c5b31041fcfd256ff6884ea7b28fb")
	fmt.Println()
}
