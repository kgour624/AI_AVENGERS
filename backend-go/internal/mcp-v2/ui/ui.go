// Package ui — MCP-UI helpers (5&6): externalUrl + renderData + postMessage bridge.
package ui

import (
	"fmt"
	"net/url"
)

// BuildIframeURL builds externalUrl via new URL(path, baseUrl).toString() (MCP 5&6 spec).
// baseUrl is from agent.fetch props — never hardcode localhost.
func BuildIframeURL(baseURL, path string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid baseUrl: %w", err)
	}
	rel, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).String(), nil
}

// UIResource is what createUIResource returns — ui:// + externalUrl.
type UIResource struct {
	URI      string `json:"uri"` // ui://expert-view/123
	URL      string `json:"iframeUrl"`
	Encoding string `json:"encoding"` // text
}

// CreateUIResource helper mirrors SDK createUIResource.
func CreateUIResource(uri, iframeURL string) UIResource {
	return UIResource{URI: uri, URL: iframeURL, Encoding: "text"}
}