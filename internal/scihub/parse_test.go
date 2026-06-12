package scihub

import "testing"

func TestExtractPDFURLFromHTML(t *testing.T) {
	t.Parallel()

	base := "https://sci-hub.st/10.1038/nature12373"

	tests := []struct {
		name    string
		html    string
		baseURL string
		want    string
	}{
		{
			name: "citation_pdf_url meta",
			html: `<html><head><meta name="citation_pdf_url" content="https://sci-hub.st/storage/2024/abc/paper.pdf"></head></html>`,
			want: "https://sci-hub.st/storage/2024/abc/paper.pdf",
		},
		{
			name: "object data storage path",
			html: `<html><body><object data="/storage/2023/bucket/hash/file.pdf"></object></body></html>`,
			want: "https://sci-hub.st/storage/2023/bucket/hash/file.pdf",
		},
		{
			name: "download link storage path",
			html: `<html><body><a href="/storage/2022/bucket/hash/file.pdf">download</a></body></html>`,
			want: "https://sci-hub.st/storage/2022/bucket/hash/file.pdf",
		},
		{
			name: "location href relative pdf",
			html: `<script>location.href='/storage/2021/bucket/hash/file.pdf';</script>`,
			want: "https://sci-hub.st/storage/2021/bucket/hash/file.pdf",
		},
		{
			name: "protocol relative embed src",
			html: `<embed src="//sci-hub.st/storage/2020/bucket/hash/file.pdf">`,
			want: "https://sci-hub.st/storage/2020/bucket/hash/file.pdf",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			baseURL := tc.baseURL
			if baseURL == "" {
				baseURL = base
			}
			got, ok := ExtractPDFURLFromHTML(tc.html, baseURL)
			if !ok {
				t.Fatal("expected PDF URL, got none")
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeDOI(t *testing.T) {
	t.Parallel()
	got := NormalizeDOI("https://doi.org/10.1038/Nature12373")
	want := "10.1038/nature12373"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseMirrorLinks(t *testing.T) {
	t.Parallel()
	html := `
		<a href="https://sci-hub.st/">ST</a>
		<a href="https://sci-hub.ru">RU</a>
		<a href="https://sci-hub.st/">dup</a>
		<a href="https://example.com/">skip</a>
	`
	got := parseMirrorLinks(html)
	want := []string{"https://sci-hub.ru", "https://sci-hub.st"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
