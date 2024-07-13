package tenancy

import (
	"context"
)

// SetTenantID is a helper function for testing to set the tenant ID in the context
func SetTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenantID)
}
