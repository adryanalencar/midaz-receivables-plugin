package command

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/sumup"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/crypto"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateAcquirerConnectionHandler_Execute(t *testing.T) {
	rawKey := "0123456789abcdef0123456789abcdef"
	encryptor, err := crypto.NewGCMEncryptor(rawKey)
	require.NoError(t, err)

	// Mock SumUp server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer valid_key" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	factory := providers.NewFactory()
	factory.Register(sumup.NewAdapter(ts.URL, ts.Client()))

	t.Run("Success with AES-256-GCM Encryption", func(t *testing.T) {
		repo := new(MockRepository)

		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_123").
			Return(&domain.ReceivablesAccount{ID: "acc_123", AccountAlias: "@merchant/receivables"}, nil)

		var savedConn *domain.AcquirerConnection
		repo.On("CreateAcquirerConnection", mock.Anything, mock.MatchedBy(func(conn *domain.AcquirerConnection) bool {
			savedConn = conn
			return conn.ReceivablesAccountID == "acc_123" &&
				conn.AcquirerType == domain.AcquirerSumUp &&
				len(conn.CredentialsEncrypted) > 0
		})).Return(nil)

		handler := NewCreateAcquirerConnectionHandler(repo, factory, encryptor)

		conn, err := handler.Execute(context.Background(), CreateAcquirerConnectionInput{
			ReceivablesAccountID: "acc_123",
			Acquirer:             domain.AcquirerSumUp,
			Credentials: domain.ConnectionCredentials{
				APIKey:       "valid_key",
				MerchantCode: "M100",
			},
			WebhookSecret: "wh_sec_1",
		})

		require.NoError(t, err)
		assert.Equal(t, "acc_123", conn.ReceivablesAccountID)
		assert.Equal(t, domain.ConnectionStatusActive, conn.Status)

		// Verify that credentials stored are encrypted and NOT in plaintext
		assert.NotContains(t, savedConn.CredentialsEncrypted, "valid_key")

		// Verify decryption in runtime restores original credentials
		decrypted, err := encryptor.Decrypt(savedConn.CredentialsEncrypted)
		require.NoError(t, err)

		var creds domain.ConnectionCredentials
		err = json.Unmarshal(decrypted, &creds)
		require.NoError(t, err)
		assert.Equal(t, "valid_key", creds.APIKey)
		assert.Equal(t, "M100", creds.MerchantCode)

		repo.AssertExpectations(t)
	})

	t.Run("Missing Parent Account", func(t *testing.T) {
		repo := new(MockRepository)
		repo.On("FindReceivablesAccountByID", mock.Anything, "missing_acc").
			Return(nil, nil)

		handler := NewCreateAcquirerConnectionHandler(repo, factory, encryptor)

		_, err := handler.Execute(context.Background(), CreateAcquirerConnectionInput{
			ReceivablesAccountID: "missing_acc",
			Acquirer:             domain.AcquirerSumUp,
			Credentials:          domain.ConnectionCredentials{APIKey: "valid_key"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "receivables account missing_acc not found")
	})

	t.Run("Invalid Credentials Rejection", func(t *testing.T) {
		repo := new(MockRepository)
		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_123").
			Return(&domain.ReceivablesAccount{ID: "acc_123"}, nil)

		handler := NewCreateAcquirerConnectionHandler(repo, factory, encryptor)

		_, err := handler.Execute(context.Background(), CreateAcquirerConnectionInput{
			ReceivablesAccountID: "acc_123",
			Acquirer:             domain.AcquirerSumUp,
			Credentials:          domain.ConnectionCredentials{APIKey: "bad_key"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential validation failed")
	})
}
