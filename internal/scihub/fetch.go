package scihub

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const minPDFBytes = 1024

// ResolvePDFURL requests the Sci-Hub landing page for a DOI and returns the PDF URL.
func ResolvePDFURL(ctx context.Context, httpClient *http.Client, mirrorBase, doi string) (string, error) {
	if httpClient == nil {
		httpClient = defaultClient
	}
	doi = NormalizeDOI(doi)
	if doi == "" {
		return "", fmt.Errorf("empty DOI")
	}
	landingURL, err := joinMirrorDOI(mirrorBase, doi)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, landingURL, nil)
	if err != nil {
		return "", fmt.Errorf("building landing request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting landing page: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("reading landing page: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("landing page returned HTTP %d", resp.StatusCode)
	}
	if isPDF(body) {
		return resp.Request.URL.String(), nil
	}

	baseURL := resp.Request.URL.String()
	if pdfURL, ok := ExtractPDFURLFromHTML(string(body), baseURL); ok {
		return pdfURL, nil
	}
	return "", fmt.Errorf("no PDF URL found in landing page HTML")
}

// DownloadPDF downloads a PDF and validates the payload before writing destPath.
func DownloadPDF(ctx context.Context, httpClient *http.Client, pdfURL, destPath string) error {
	if httpClient == nil {
		httpClient = defaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pdfURL, nil)
	if err != nil {
		return fmt.Errorf("building PDF request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading PDF: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("PDF download returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return fmt.Errorf("reading PDF body: %w", err)
	}
	if len(data) < minPDFBytes {
		return fmt.Errorf("PDF too small (%d bytes)", len(data))
	}
	if !isPDF(data) {
		return fmt.Errorf("response is not a PDF")
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("writing PDF: %w", err)
	}
	return nil
}

// FetchDOI resolves and downloads a PDF for doi into destPath.
func FetchDOI(ctx context.Context, httpClient *http.Client, doi, destPath, mirrorOverride string) (string, error) {
	if httpClient == nil {
		httpClient = defaultClient
	}
	mirror, err := pickMirror(ctx, mirrorOverride)
	if err != nil {
		return "", err
	}
	pdfURL, err := ResolvePDFURL(ctx, httpClient, mirror, doi)
	if err != nil {
		return "", err
	}
	if err := DownloadPDF(ctx, httpClient, pdfURL, destPath); err != nil {
		return "", err
	}
	return mirror, nil
}

func pickMirror(ctx context.Context, mirrorOverride string) (string, error) {
	candidates := []string{}
	if mirrorOverride != "" {
		candidates = append(candidates, strings.TrimRight(strings.TrimSpace(mirrorOverride), "/"))
	}
	if envMirror := strings.TrimSpace(os.Getenv("SCIHUB_MIRROR")); envMirror != "" {
		candidates = append(candidates, strings.TrimRight(envMirror, "/"))
	}
	mirrors, err := ListMirrors(ctx)
	if err == nil {
		candidates = append(candidates, mirrors...)
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no Sci-Hub mirrors available; set SCIHUB_MIRROR")
	}
	seen := make(map[string]struct{})
	for _, mirror := range candidates {
		mirror = strings.TrimRight(strings.TrimSpace(mirror), "/")
		if mirror == "" {
			continue
		}
		if _, ok := seen[mirror]; ok {
			continue
		}
		seen[mirror] = struct{}{}
		return mirror, nil
	}
	return "", fmt.Errorf("no Sci-Hub mirrors available; set SCIHUB_MIRROR")
}

func joinMirrorDOI(mirrorBase, doi string) (string, error) {
	mirrorBase = strings.TrimRight(strings.TrimSpace(mirrorBase), "/")
	if mirrorBase == "" {
		return "", fmt.Errorf("empty mirror base URL")
	}
	base, err := url.Parse(mirrorBase)
	if err != nil {
		return "", fmt.Errorf("invalid mirror base URL: %w", err)
	}
	ref, err := url.Parse(NormalizeDOI(doi))
	if err != nil {
		return "", fmt.Errorf("invalid DOI: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}

func isPDF(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "%PDF"
}
