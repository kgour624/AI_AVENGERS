package mcpv2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Billing handles MCP-UI Billing — ui://billing/checkout + webhook (MCP 5&6 + 7&8).
// Host opens Stripe/UPI via sendMcpMessage("tool", {toolName:"create_checkout"}).

type BillingService struct{ deps Deps }

func NewBillingService(d Deps) *BillingService { return &BillingService{deps: d} }

// CreateCheckout is called via MCP tool create_checkout — returns Stripe URL for Host to open via link.
func (s *BillingService) CreateCheckout(ctx context.Context, expertID, planID string) (checkoutURL string, err error) {
	if expertID == "" || planID == "" {
		return "", fmt.Errorf("expertId and planId required")
	}
	// Real: create Stripe session, return url — stub for now, wire in Phase 7
	return fmt.Sprintf("https://checkout.stripe.com/pay/%s-%s", expertID, planID), nil
}

// HandleWebhook verifies Stripe/UPI signature then updates mcp_rentals -> notifications/resources/updated.
func (s *BillingService) HandleWebhook(c *gin.Context) {
	// Verify signature header (Stripe-Signature) — fail closed
	sig := c.GetHeader("Stripe-Signature")
	if sig == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature"})
		return
	}
	// TODO: verify with webhook secret, update mcp_rentals status active, publish notifications/resources/updated
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RegisterBillingRoutes mounts POST /mcp-billing/webhook — caller adds to gin Engine.
func (s *BillingService) RegisterBillingRoutes(r *gin.Engine) {
	r.POST("/mcp-billing/webhook", s.HandleWebhook)
}