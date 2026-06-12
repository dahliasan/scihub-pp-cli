package scihub

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const mirrorsPageURL = "https://www.sci-hub.pub/"

var (
	hrefRe        = regexp.MustCompile(`(?is)href\s*=\s*["']([^"']+)["']`)
	mirrorHostRe  = regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.)?(sci-hub\.[a-z0-9.-]+)/?$`)
	defaultClient = NewHTTPClient(30 * time.Second)
)

// ListMirrors fetches sci-hub.pub and returns deduplicated mirror base URLs.
func ListMirrors(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mirrorsPageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building mirrors request: %w", err)
	}
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching mirrors page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("mirrors page returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("reading mirrors page: %w", err)
	}
	return parseMirrorLinks(string(body)), nil
}

func parseMirrorLinks(html string) []string {
	seen := make(map[string]struct{})
	var mirrors []string
	for _, match := range hrefRe.FindAllStringSubmatch(html, -1) {
		if len(match) < 2 {
			continue
		}
		base, ok := normalizeMirrorURL(match[1])
		if !ok {
			continue
		}
		if _, exists := seen[base]; exists {
			continue
		}
		seen[base] = struct{}{}
		mirrors = append(mirrors, base)
	}
	sort.Strings(mirrors)
	return mirrors
}

func normalizeMirrorURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	if !strings.Contains(strings.ToLower(raw), "sci-hub.") {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if !mirrorHostRe.MatchString(host) && !strings.HasPrefix(host, "sci-hub.") {
		return "", false
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return strings.TrimRight(fmt.Sprintf("%s://%s", scheme, host), "/"), true
}
