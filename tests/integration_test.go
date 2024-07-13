package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/config"
	"github.com/api-moose/company-earnings/internal/middleware/access_control"
	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/pocketbase/pocketbase/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthApp struct {
	mock.Mock
}

func (m *MockAuthApp) FindAuthRecordByToken(token, secret string) (*models.Record, error) {
	args := m.Called(token, secret)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Record), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthApp) GetAuthTokenSecret() string {
	args := m.Called()
	return args.String(0)
}

func createIntegrationMockRecord() *models.Record {
	collection := &models.Collection{
		BaseModel: models.BaseModel{
			Id: "mockCollectionId",
		},
		Name: "mockCollection",
	}

	record := models.NewRecord(collection)
	record.SetId("mockId")
	record.Set("role", "admin")
	record.Set("tenantId", "test-tenant")
	return record
}

func TestMiddlewareIntegration(t *testing.T) {
	mockAuth := new(MockAuthApp)
	mockRecord := createIntegrationMockRecord()

	mockAuth.On("FindAuthRecordByToken", "valid-token", "test_secret").Return(mockRecord, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Apply middleware chain
	handlerWithMiddleware := auth.AuthMiddleware(appConfig.Adapter)(
		tenancy.TenantMiddleware(
			access_control.RBACMiddleware(handler),
		),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-Tenant-ID", "test-tenant")
	rr := httptest.NewRecorder()

	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Success", rr.Body.String())
}

func TestContextPropagation(t *testing.T) {
	mockAuth := new(MockAuthApp)
	mockRecord := createIntegrationMockRecord()

	mockAuth.On("FindAuthRecordByToken", "valid-token", "test_secret").Return(mockRecord, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, userOk := r.Context().Value("user").(*models.Record)
		tenantID, tenantOk := tenancy.GetTenantID(r)

		if !userOk || !tenantOk || user == nil || tenantID == "" {
			http.Error(w, "Context propagation failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Context propagated successfully"))
	})

	// Apply middleware chain
	handlerWithMiddleware := auth.AuthMiddleware(appConfig.Adapter)(
		tenancy.TenantMiddleware(
			access_control.RBACMiddleware(handler),
		),
	)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-Tenant-ID", "test-tenant")
	rr := httptest.NewRecorder()

	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Context propagated successfully", rr.Body.String())
}
