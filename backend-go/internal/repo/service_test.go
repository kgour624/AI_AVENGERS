package repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func httpResponse(status int, headers http.Header, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     headers,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestFetchGitLabTreeFollowsPagination(t *testing.T) {
	var requestedPages []string
	svc := &Service{
		logger: zap.NewNop(),
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			page := req.URL.Query().Get("page")
			requestedPages = append(requestedPages, page)
			switch page {
			case "1":
				headers := make(http.Header)
				headers.Set("X-Next-Page", "2")
				return httpResponse(http.StatusOK, headers, `[
					{"path":"cmd/app/main.go","type":"blob"},
					{"path":"internal","type":"tree"},
					{"path":"README.txt","type":"blob"}
				]`), nil
			case "2":
				return httpResponse(http.StatusOK, make(http.Header), `[
					{"path":"internal/store/store.go","type":"blob"}
				]`), nil
			default:
				return nil, fmt.Errorf("unexpected page %q", page)
			}
		})},
	}

	entries, err := svc.fetchGitLabTree(context.Background(), "https://gitlab.com/acme/service", "main", "token")
	if err != nil {
		t.Fatalf("fetchGitLabTree() error = %v", err)
	}
	// The tree now keeps every blob (unsupported ones included) so the stored
	// structure is complete; only content fetching is filtered later.
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.Path)
	}
	if want := []string{"cmd/app/main.go", "README.txt", "internal/store/store.go"}; fmt.Sprint(paths) != fmt.Sprint(want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	if want := []string{"1", "2"}; fmt.Sprint(requestedPages) != fmt.Sprint(want) {
		t.Fatalf("requested pages = %v, want %v", requestedPages, want)
	}
}

func TestFetchGitLabTreeRejectsHTTPError(t *testing.T) {
	svc := &Service{
		logger: zap.NewNop(),
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusUnauthorized, make(http.Header), `{"message":"unauthorized"}`), nil
		})},
	}
	if _, err := svc.fetchGitLabTree(context.Background(), "https://gitlab.com/acme/service", "main", "bad-token"); err == nil {
		t.Fatal("fetchGitLabTree() expected an error for HTTP 401")
	}
}

func TestFetchGitLabTreeEnforcesPageCap(t *testing.T) {
	requests := 0
	svc := &Service{
		logger: zap.NewNop(),
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requests++
			headers := make(http.Header)
			headers.Set("X-Next-Page", fmt.Sprintf("%d", requests+1))
			return httpResponse(http.StatusOK, headers, `[]`), nil
		})},
	}
	if _, err := svc.fetchGitLabTree(context.Background(), "https://gitlab.com/acme/service", "main", "token"); err != nil {
		t.Fatalf("fetchGitLabTree() error = %v", err)
	}
	if requests != gitLabMaxTreePages {
		t.Fatalf("requests = %d, want page cap %d", requests, gitLabMaxTreePages)
	}
}

func TestReadCappedFileSkipsOversizedContent(t *testing.T) {
	svc := &Service{logger: zap.NewNop()}
	oversized := strings.Repeat("x", maxRepoFileBytes+1)
	content, err := svc.readCappedFile(strings.NewReader(oversized), "large.go")
	if err == nil {
		t.Fatal("readCappedFile() expected oversized content to be rejected")
	}
	if content != "" {
		t.Fatalf("readCappedFile() returned partial content of %d bytes", len(content))
	}
}

func TestFetchGitLabFileContentRejectsHTTPError(t *testing.T) {
	svc := &Service{
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusNotFound, make(http.Header), `{"message":"not found"}`), nil
		})},
	}
	content, err := svc.fetchFileContent(context.Background(), ProviderGitLab, "https://gitlab.com/acme/service", "main", "missing.go", "token")
	if err == nil || content != "" {
		t.Fatalf("fetchFileContent() = (%q, %v), want empty content and HTTP error", content, err)
	}
}

func TestExchangeOAuthCodeRetainsGitLabRefreshTokenAndExpiry(t *testing.T) {
	svc := &Service{
		gitlabOAuth: OAuthConfig{ClientID: "client", ClientSecret: "secret", RedirectURL: "https://app.example/callback"},
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK, make(http.Header), `{"access_token":"access","refresh_token":"refresh","expires_in":7200}`), nil
		})},
	}

	tok, err := svc.ExchangeOAuthCode(context.Background(), ProviderGitLab, "code")
	if err != nil {
		t.Fatalf("ExchangeOAuthCode() error = %v", err)
	}
	if tok.AccessToken != "access" || tok.RefreshToken != "refresh" {
		t.Fatalf("token set = %+v, want access and refresh tokens", tok)
	}
	if remaining := time.Until(tok.ExpiresAt); remaining < 7190*time.Second || remaining > 7200*time.Second {
		t.Fatalf("expiry remaining = %s, want close to 2h", remaining)
	}
}

func TestRefreshGitLabTokenUsesRefreshGrant(t *testing.T) {
	svc := &Service{
		gitlabOAuth: OAuthConfig{ClientID: "client", ClientSecret: "secret", RedirectURL: "https://app.example/callback"},
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				return nil, err
			}
			if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "old-refresh" {
				return nil, fmt.Errorf("unexpected refresh request form: %v", form)
			}
			return httpResponse(http.StatusOK, make(http.Header), `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":7200}`), nil
		})},
	}

	tok, err := svc.refreshGitLabToken(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("refreshGitLabToken() error = %v", err)
	}
	if tok.AccessToken != "new-access" || tok.RefreshToken != "new-refresh" || tok.ExpiresAt.IsZero() {
		t.Fatalf("refreshed token = %+v, missing refreshed credentials/expiry", tok)
	}
}

func TestGitLabTreeEntriesCarryNoSize(t *testing.T) {
	svc := &Service{
		logger: zap.NewNop(),
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK, make(http.Header), `[{"path":"a.go","type":"blob"}]`), nil
		})},
	}
	entries, err := svc.fetchGitLabTree(context.Background(), "https://gitlab.com/acme/service", "main", "token")
	if err != nil {
		t.Fatalf("fetchGitLabTree() error = %v", err)
	}
	if len(entries) != 1 || entries[0].SizeBytes != nil || entries[0].BlobSHA != "" {
		t.Fatalf("entries = %+v, want one entry with nil size and no provider blob SHA", entries)
	}
}

func TestSupportedFilesDropsUnsupportedAndOversized(t *testing.T) {
	big := maxRepoFileBytes + 1
	entries := []repoTreeItem{
		{Path: "main.go"},
		{Path: "logo.png"},
		{Path: "huge.go", SizeBytes: &big},
		{Path: "unknown_size.go"},
	}
	got := supportedFiles(entries)
	want := []string{"main.go", "unknown_size.go"}
	paths := make([]string, 0, len(got))
	for _, entry := range got {
		paths = append(paths, entry.Path)
	}
	if fmt.Sprint(paths) != fmt.Sprint(want) {
		t.Fatalf("supportedFiles() = %v, want %v", paths, want)
	}
}

func TestGitBlobSHAMatchesGitBlobHash(t *testing.T) {
	// git hash-object of "hello\n" is a known value; this pins the
	// "blob <len>\x00" framing, which is the entire reason the SHA is computed
	// locally instead of trusted from a provider field.
	if got, want := gitBlobSHA("hello\n"), "ce013625030ba8dba906f756967f9e9ca394464a"; got != want {
		t.Fatalf("gitBlobSHA() = %s, want %s", got, want)
	}
}

func TestFetchHeadCommitSHAParsesBothProviders(t *testing.T) {
	github := &Service{
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK, make(http.Header), `{"sha":"abc123"}`), nil
		})},
	}
	sha, err := github.fetchHeadCommitSHA(context.Background(), ProviderGitHub, "https://github.com/acme/service", "main", "token")
	if err != nil || sha != "abc123" {
		t.Fatalf("github head sha = (%q, %v), want abc123", sha, err)
	}

	gitlab := &Service{
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK, make(http.Header), `[{"id":"def456"}]`), nil
		})},
	}
	sha, err = gitlab.fetchHeadCommitSHA(context.Background(), ProviderGitLab, "https://gitlab.com/acme/service", "main", "token")
	if err != nil || sha != "def456" {
		t.Fatalf("gitlab head sha = (%q, %v), want def456", sha, err)
	}
}

func TestFetchHeadCommitSHAErrorsOnEmptyResult(t *testing.T) {
	svc := &Service{
		httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return httpResponse(http.StatusOK, make(http.Header), `[]`), nil
		})},
	}
	if _, err := svc.fetchHeadCommitSHA(context.Background(), ProviderGitLab, "https://gitlab.com/acme/service", "main", "token"); err == nil {
		t.Fatal("fetchHeadCommitSHA() expected an error for an empty commit list")
	}
}

// The one-click OAuth connect must refuse before it builds any URL when the
// deployment has no OAuth app credentials. Previously the authorize URL was
// assembled with an empty client_id and the browser landed on the provider's
// error page, which reads as "the button is broken" rather than "the server is
// not configured".
func TestOAuthCredentialsMissing(t *testing.T) {
	configured := &Service{
		githubOAuth: OAuthConfig{ClientID: "gh-id", ClientSecret: "gh-secret"},
		gitlabOAuth: OAuthConfig{ClientID: "gl-id", ClientSecret: "gl-secret"},
	}
	unconfigured := &Service{}

	tests := []struct {
		name     string
		svc      *Service
		provider string
		want     bool
		wantVars string
	}{
		{name: "github configured", svc: configured, provider: ProviderGitHub, want: false},
		{name: "gitlab configured", svc: configured, provider: ProviderGitLab, want: false},
		{
			name: "github missing", svc: unconfigured, provider: ProviderGitHub,
			want: true, wantVars: "GITHUB_CLIENT_ID/GITHUB_CLIENT_SECRET",
		},
		{
			name: "gitlab missing", svc: unconfigured, provider: ProviderGitLab,
			want: true, wantVars: "GITLAB_CLIENT_ID/GITLAB_CLIENT_SECRET",
		},
		{
			// Half-configured is still unconfigured: a client id with no secret
			// cannot complete the token exchange, so the handshake must not start.
			name:     "github id without secret",
			svc:      &Service{githubOAuth: OAuthConfig{ClientID: "gh-id"}},
			provider: ProviderGitHub, want: true, wantVars: "GITHUB_CLIENT_ID/GITHUB_CLIENT_SECRET",
		},
		{
			// An unknown provider is not reported here — GetOAuthURL answers that
			// with "unsupported provider", and claiming a missing credential for a
			// provider we do not support would point the operator at the wrong fix.
			name: "unknown provider", svc: unconfigured, provider: "bitbucket", want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			missing, vars := tc.svc.oauthCredentialsMissing(tc.provider)
			if missing != tc.want {
				t.Fatalf("missing = %v, want %v", missing, tc.want)
			}
			if missing && vars != tc.wantVars {
				t.Fatalf("env vars = %q, want %q", vars, tc.wantVars)
			}
		})
	}
}

func TestGetOAuthURLRefusesWithoutCredentials(t *testing.T) {
	s := &Service{}
	if _, err := s.GetOAuthURL(context.Background(), ProviderGitHub, "state", "project"); !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("GetOAuthURL error = %v, want ErrOAuthNotConfigured", err)
	}
}
