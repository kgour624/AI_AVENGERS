package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

func TestCodeCraftAPIRefreshPricingUsesPerModelCatalogRates(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %q, want /v1/models", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q, want bearer key", got)
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"strong-model","pricing":{"input_per_1k":0.002,"output_per_1k":0.009}},{"id":"cheap-model","pricing":{"input_per_1k":0.0001,"output_per_1k":0.0004}}]}`))
	}))
	defer server.Close()

	provider := NewCodeCraftAPIProvider("test-key", server.URL+"/v1", "cheap-model", "strong-model", "", server.Client())
	if in, out := provider.CostPer1K(gtypes.ModelStrong); in != 0 || out != 0 {
		t.Fatalf("cost before catalog fetch = (%v, %v), want unknown zero", in, out)
	}
	if err := provider.RefreshPricing(context.Background()); err != nil {
		t.Fatalf("RefreshPricing() error = %v", err)
	}
	if in, out := provider.CostPer1K(gtypes.ModelStrong); in != 0.002 || out != 0.009 {
		t.Fatalf("strong cost = (%v, %v), want (0.002, 0.009)", in, out)
	}
	if in, out := provider.CostPer1K(gtypes.ModelCheap); in != 0.0001 || out != 0.0004 {
		t.Fatalf("cheap cost = (%v, %v), want (0.0001, 0.0004)", in, out)
	}
	if in, out := provider.CostPer1K(gtypes.ModelFast); in != 0 || out != 0 {
		t.Fatalf("unlisted model cost = (%v, %v), want unknown zero", in, out)
	}
	if err := provider.RefreshPricing(context.Background()); err != nil {
		t.Fatalf("cached RefreshPricing() error = %v", err)
	}
	if requests != 1 {
		t.Fatalf("catalog requests = %d, want one within the cache window", requests)
	}
}

func TestCodeCraftAPIRefreshPricingPreservesLastKnownRatesOnFailure(t *testing.T) {
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != http.StatusOK {
			http.Error(w, "unavailable", status)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"strong-model","pricing":{"input_per_1k":0.002,"output_per_1k":0.009}}]}`))
	}))
	defer server.Close()

	provider := NewCodeCraftAPIProvider("test-key", server.URL+"/v1", "", "strong-model", "", server.Client())
	if err := provider.RefreshPricing(context.Background()); err != nil {
		t.Fatalf("initial RefreshPricing() error = %v", err)
	}
	provider.pricingMu.Lock()
	provider.pricingTriedAt = time.Time{}
	provider.pricingMu.Unlock()

	status = http.StatusBadGateway
	if err := provider.RefreshPricing(context.Background()); err == nil {
		t.Fatal("RefreshPricing() should report an upstream error")
	}
	if in, out := provider.CostPer1K(gtypes.ModelStrong); in != 0.002 || out != 0.009 {
		t.Fatalf("retained cost = (%v, %v), want previous rates (0.002, 0.009)", in, out)
	}
}
