package internal

import (
	"net/url"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

type LanguageDomainTarget struct {
	Scheme         string
	Host           string
	Path           string
	NormalizedHost string
}

func NormalizeHost(hostName string) string {
	hostName = strings.ToLower(strings.TrimSpace(hostName))
	if hostName == "" {
		return ""
	}
	// Fast path: simple hostname without scheme or port
	if !strings.Contains(hostName, "://") && !strings.Contains(hostName, ":") && !strings.Contains(hostName, "/") {
		return strings.TrimPrefix(hostName, "www.")
	}
	// Strip port if present (e.g., "example.com:8080")
	if idx := strings.LastIndex(hostName, ":"); idx > 0 && !strings.Contains(hostName, "://") {
		hostName = hostName[:idx]
		return strings.TrimPrefix(hostName, "www.")
	}
	// Parse full URL
	parsed, err := url.Parse(hostName)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("https://" + hostName)
		if err != nil {
			return strings.TrimPrefix(hostName, "www.")
		}
	}
	host := parsed.Hostname()
	if host == "" {
		host = hostName
	}
	return strings.TrimPrefix(host, "www.")
}

func ParseLanguageDomainTarget(raw string) (*LanguageDomainTarget, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("https://" + raw)
		if err != nil {
			return nil, false
		}
	}
	target := &LanguageDomainTarget{
		Scheme: "https://",
		Path:   "/",
	}
	if parsed.Scheme != "" {
		target.Scheme = parsed.Scheme + "://"
	}
	if parsed.Hostname() != "" {
		target.Host = parsed.Hostname()
	} else {
		target.Host = NormalizeHost(raw)
	}
	if parsed.Path != "" {
		target.Path = parsed.Path
	}
	target.NormalizedHost = NormalizeHost(target.Host)
	return target, true
}

func getLanguageDomainTarget(cfg *types.Config, key string) (*LanguageDomainTarget, bool) {
	if cfg == nil || len(cfg.LanguageDomains) == 0 {
		return nil, false
	}
	raw := strings.TrimSpace(cfg.LanguageDomains[key])
	if raw == "" {
		return nil, false
	}
	return ParseLanguageDomainTarget(raw)
}

func GetDefaultLanguageDomainTarget(cfg *types.Config) (*LanguageDomainTarget, bool) {
	if cfg == nil {
		return nil, false
	}
	if target, ok := getLanguageDomainTarget(cfg, "default"); ok {
		return target, true
	}
	if cfg.General.DefaultLanguage != "" {
		if target, ok := getLanguageDomainTarget(cfg, cfg.General.DefaultLanguage); ok {
			return target, true
		}
	}
	if cfg.General.CanonicalUrl != "" {
		if target, ok := ParseLanguageDomainTarget(cfg.General.CanonicalUrl); ok {
			return target, true
		}
	}
	return nil, false
}

func HostInLanguageDomains(host string, domains map[string]string) bool {
	if len(domains) == 0 || host == "" {
		return false
	}
	normalized := NormalizeHost(host)
	for _, raw := range domains {
		// Fast path: just normalize the domain value without full parsing
		if NormalizeHost(raw) == normalized {
			return true
		}
	}
	return false
}

// GetLanguageForDomain returns the language code for which this domain is configured.
// Returns empty string if domain is the default domain or not found in language_domains.
// NOTE: If multiple languages share the same domain, returns the first one found (non-deterministic).
// Use IsDomainConfiguredForLanguage for checking specific language.
func GetLanguageForDomain(host string, cfg *types.Config) string {
	if cfg == nil || len(cfg.LanguageDomains) == 0 || host == "" {
		return ""
	}
	normalized := NormalizeHost(host)
	defaultTarget, hasDefault := GetDefaultLanguageDomainTarget(cfg)
	// If this is the default domain, return empty
	if hasDefault && defaultTarget.NormalizedHost == normalized {
		return ""
	}
	// Find which language this domain belongs to
	for lang, raw := range cfg.LanguageDomains {
		if lang == "default" {
			continue
		}
		if NormalizeHost(raw) == normalized {
			return lang
		}
	}
	return ""
}

// IsDomainConfiguredForLanguage checks if the given domain is configured for the specified language.
// Handles cases where multiple languages share the same domain.
func IsDomainConfiguredForLanguage(host, langId string, cfg *types.Config) bool {
	if cfg == nil || len(cfg.LanguageDomains) == 0 || host == "" || langId == "" {
		return false
	}
	langDomain, ok := cfg.LanguageDomains[langId]
	if !ok || langDomain == "" {
		return false
	}
	return NormalizeHost(langDomain) == NormalizeHost(host)
}

func GetDefaultLanguageDomainValue(cfg *types.Config) (string, bool) {
	if cfg == nil || len(cfg.LanguageDomains) == 0 {
		if cfg != nil && cfg.General.CanonicalUrl != "" {
			if parsed, err := url.Parse(cfg.General.CanonicalUrl); err == nil && parsed.Host != "" {
				return parsed.Hostname(), true
			}
		}
		return "", false
	}
	if v := strings.TrimSpace(cfg.LanguageDomains["default"]); v != "" {
		return v, true
	}
	if cfg.General.DefaultLanguage != "" {
		if v := strings.TrimSpace(cfg.LanguageDomains[cfg.General.DefaultLanguage]); v != "" {
			return v, true
		}
	}
	if cfg.General.CanonicalUrl != "" {
		if parsed, err := url.Parse(cfg.General.CanonicalUrl); err == nil && parsed.Host != "" {
			return parsed.Hostname(), true
		}
	}
	return "", false
}
