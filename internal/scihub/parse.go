package scihub

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	citationPDFMetaRe = regexp.MustCompile(`(?is)<meta[^>]+name\s*=\s*["']citation_pdf_url["'][^>]+content\s*=\s*["']([^"']+)["']`)
	citationPDFMetaRe2 = regexp.MustCompile(`(?is)<meta[^>]+content\s*=\s*["']([^"']+)["'][^>]+name\s*=\s*["']citation_pdf_url["']`)
	objectDataRe       = regexp.MustCompile(`(?is)<object[^>]+data\s*=\s*["']([^"']+)["']`)
	storageLinkRe      = regexp.MustCompile(`(?is)(?:href|src)\s*=\s*["'](/storage/[^"']+)["']`)
	locationHrefRe     = regexp.MustCompile(`(?is)location\.href\s*=\s*["']([^"']+)["']`)
	embedSrcRe         = regexp.MustCompile(`(?is)<embed[^>]+src\s*=\s*["']([^"']+)["']`)
	iframeSrcRe        = regexp.MustCompile(`(?is)<iframe[^>]+src\s*=\s*["']([^"']+)["']`)
)

// ExtractPDFURLFromHTML parses Sci-Hub landing HTML and returns an absolute PDF URL.
func ExtractPDFURLFromHTML(html string, baseURL string) (string, bool) {
	candidates := []string{
		firstSubmatch(citationPDFMetaRe, html),
		firstSubmatch(citationPDFMetaRe2, html),
		firstSubmatch(objectDataRe, html),
		firstSubmatch(storageLinkRe, html),
		firstSubmatch(locationHrefRe, html),
		firstSubmatch(embedSrcRe, html),
		firstSubmatch(iframeSrcRe, html),
	}
	for _, raw := range candidates {
		if raw == "" {
			continue
		}
		resolved, err := resolveAgainstBase(raw, baseURL)
		if err != nil {
			continue
		}
		return resolved, true
	}
	return "", false
}

func firstSubmatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func resolveAgainstBase(rawURL, baseURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty URL")
	}
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}
