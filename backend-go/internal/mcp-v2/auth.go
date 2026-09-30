package mcpv2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthInfo is piped via props (MCP 7&8 — AuthInfo pattern).
type AuthInfo struct {
	Token    string
	ClientID string
	Scopes   []string
	UserID   string
}

// SUPPORTED_SCOPES — single source of truth for MCP-V2.
var SUPPORTED_SCOPES = []string{"expert:read", "chunk:read", "expert:write", "admin:write", "billing:read"}

// hasScopes checks if auth has all required scopes.
func hasScopes(auth AuthInfo, required ...string) bool {
	set := make(map[string]bool, len(auth.Scopes))
	for _, s := range auth.Scopes {
		set[s] = true
	}
	for _, r := range required {
		if !set[r] {
			return false
		}
	}
	return true
}

// validateScopes returns 403 if missing.
func validateScopes(c *gin.Context, auth AuthInfo, required ...string) bool {
	if !hasScopes(auth, required...) {
		c.Header("WWW-Authenticate", fmt.Sprintf(`Bearer error="insufficient_scope", error_description="Need one of: %s"`, strings.Join(required, ",")))
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient_scope", "required": required})
		return false
	}
	return true
}

// resolveAuthInfo introspects Bearer token via POST /oauth/introspect (MCP 7&8 spec).
// Uses application/x-www-form-urlencoded body `token`, strips "Bearer " prefix.
func resolveAuthInfo(r *http.Request, introspectURL string) (*AuthInfo, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return nil, fmt.Errorf("missing Authorization")
	}
	token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer"))
	token = strings.TrimSpace(strings.TrimPrefix(token, "bearer"))
	// Replace regex /^Bearer\s+/i equivalent above
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}
	form := url.Values{"token": {token}}
	req, err := http.NewRequest(http.MethodPost, introspectURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspect status %d", resp.StatusCode)
	}
	var body struct {
		Active   bool   `json:"active"`
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
		Sub      string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if !body.Active {
		return nil, fmt.Errorf("token not active")
	}
	scopes := strings.Fields(body.Scope)
	return &AuthInfo{Token: token, ClientID: body.ClientID, Scopes: scopes, UserID: body.Sub}, nil
}