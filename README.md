# hashtray-go

A Go rewrite of [hashtray](https://github.com/balestek/hashtray) — an OSINT tool for Gravatar.

Find a Gravatar account from an email address, or locate an email address from a Gravatar username or hash. A single, statically-linked binary with no runtime dependencies.

## Installation

### Download binary

Download the latest binary for your platform from [Releases](https://github.com/zackey-heuristics/hashtray-Go/releases).

### Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/zackey-heuristics/hashtray-Go.git
cd hashtray-Go
make build
./hashtray-go --help
```

## Usage

### Find a Gravatar account from an email

```bash
hashtray-go email user@example.com
```

Converts the email to its MD5 hash, checks if a Gravatar profile exists, and displays the profile information.

### Find an email from a Gravatar username or hash

```bash
hashtray-go account username
hashtray-go account 437e4dc6d001f2519bc9e7a6b6412923
```

#### Options

| Flag | Description |
|------|-------------|
| `-l`, `--domain_list` | Domain list: `common` (455, default), `long` (5,334), `full` (118,062) |
| `-e`, `--elements` | Custom elements for email generation |
| `-d`, `--domains` | Custom email domains |
| `-c`, `--crazy` | Try every separator combination (exhaustive, much slower) |

#### Examples

```bash
# Use the long domain list
hashtray-go account jondo -l long

# Provide custom elements
hashtray-go account jondo -e john doe j d jondo 2001

# Use custom domains
hashtray-go account jondo -d domain1.com domain2.com
```

## Retrievable Information

If the profile is public, the following can be retrieved:

- Hash, Profile URL, Avatar
- Location, Display name, Preferred username
- Pronunciation, Pronouns, Bio
- Job title, Company
- Emails, Phone numbers, Contact info
- Verified accounts (Twitter, Instagram, LinkedIn, Bluesky, etc.)
- Payment info (PayPal, Venmo, crypto wallets)
- Photos, Interests, Links

## Credits

Based on [hashtray](https://github.com/balestek/hashtray) by balestek.

Techniques:
- [BanPangar](https://twitter.com/BanPangar/status/1357805358153150467)
- [cyb_detective](https://publication.osintambition.org/4-easy-tricks-for-using-gravatar-in-osint-99c0910d933)

## License

GPLv3
