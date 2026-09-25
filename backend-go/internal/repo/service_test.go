package repo

import (
	"context"
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

	files, err := svc.fetchGitLabTree(context.Background(), "https://gitlab.com/acme/service", "main", "token")
	if err != nil {
		t.Fatalf("fetchGitLabTree() error = %v", err)
	}
	if want := []string{"cmd/app/main.go", "internal/store/store.go"}; fmt.Sprint(files) != fmt.Sprint(want) {
		t.Fatalf("files = %v, want %v", files, want)
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
