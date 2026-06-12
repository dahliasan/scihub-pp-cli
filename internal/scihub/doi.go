package scihub

import "strings"

// NormalizeDOI strips common prefixes and lowercases the DOI string.
func NormalizeDOI(doi string) string {
	doi = strings.TrimSpace(doi)
	doi = strings.TrimPrefix(doi, "https://doi.org/")
	doi = strings.TrimPrefix(doi, "http://doi.org/")
	doi = strings.TrimPrefix(doi, "doi:")
	return strings.ToLower(doi)
}
