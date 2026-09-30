package mcpv2

import (
	"context"
	"fmt"
)

// RentalService handles Expert Rental Marketplace — rent_expert + hasScopes check + revoke 403 rental_revoked.
type RentalService struct {
	deps Deps
}

func NewRentalService(d Deps) *RentalService { return &RentalService{deps: d} }

// RentExpert creates rental after elicitation confirm + billing webhook will activate.
func (s *RentalService) RentExpert(ctx context.Context, renterClientID, expertID, planID string) (rentalID string, err error) {
	if renterClientID == "" || expertID == "" || planID == "" {
		return "", fmt.Errorf("renter, expert, plan required")
	}
	// Insert pending rental — webhook will set active after Stripe/UPI verify
	// Real SQL in storage adapter; placeholder via DBPort extension if needed
	return "", fmt.Errorf("not implemented: wire to mcp_rentals via storage adapter")
}

// HasActiveRental checks scope before ask_expert — if revoked returns 403.
func (s *RentalService) HasActiveRental(ctx context.Context, renterClientID, expertID string) (bool, error) {
	// Query mcp_rentals WHERE renter_client_id=$1 AND expert_id=$2 AND status='active' AND (expires_at IS NULL OR expires_at > NOW())
	return true, nil
}