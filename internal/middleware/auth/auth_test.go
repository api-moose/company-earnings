package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
	"github.com/rogpeppe/go-internal/cache"
)

type MockApp struct {
	RecordData *models.Record
}

func (m *MockApp) Dao() *daos.Dao {
	return &daos.Dao{}
}

func (m *MockApp) FindAuthRecordByToken(token string, baseTokenKey string) (*models.Record, error) {
	if m.RecordData != nil {
		return m.RecordData, nil
	}
	return nil, nil
}

func (m *MockApp) Settings() *core.Settings {
	return &core.Settings{
		RecordAuthToken: &core.TokenConfig{Secret: "test_secret"},
	}
}

// Implement other required methods of core.App interface
func (m *MockApp) Bootstrap() error                               { return nil }
func (m *MockApp) ResetBootstrapState() error                     { return nil }
func (m *MockApp) CreateBackup(context.Context, string) error     { return nil }
func (m *MockApp) RestoreBackup(context.Context, io.Reader) error { return nil }
func (m *MockApp) Migrate() error                                 { return nil }
func (m *MockApp) IsBootstrapped() bool                           { return true }
func (m *MockApp) RefreshSettings() error                         { return nil }
func (m *MockApp) DataDir() string                                { return "" }
func (m *MockApp) EncryptionKey() []byte                          { return nil }
func (m *MockApp) IsDev() bool                                    { return true }
func (m *MockApp) AppUrl() string                                 { return "" }
func (m *MockApp) Logger() *util.SerializedLog                    { return nil }
func (m *MockApp) SubscriptionsBroker() *subscriptions.Broker     { return nil }
func (m *MockApp) Cache() *cache.Cache                            { return nil }
func (m *MockApp) NewMailClient() mailer.Mailer                   { return nil }
func (m *MockApp) DB() *dbx.DB                                    { return nil }
func (m *MockApp) Dao() *daos.Dao                                 { return nil }
func (m *MockApp) Models() map[string]models.Model                { return nil }

func TestAuthMiddleware(t *testing.T) {
	collection := &models.Collection{
		Name: "users",
		Schema: schema.NewSchema(
			&schema.SchemaField{Name: "email", Type: schema.FieldTypeEmail},
		),
	}

	tests := []struct {
		name           string
		token          string
		record         *models.Record
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          "valid_token",
			record:         models.NewRecord(collection),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			token:          "",
			record:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			token:          "invalid_token",
			record:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record != nil {
				tt.record.Set("email", "test@example.com")
			}

			mockApp := &MockApp{RecordData: tt.record}

			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rr := httptest.NewRecorder()

			handler := AuthMiddleware(mockApp)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, ok := r.Context().Value("user").(*models.Record)
				if !ok {
					t.Error("User not found in context")
				}
				if user != tt.record {
					t.Error("User in context does not match expected user")
				}
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
