package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetDatabaseURI(t *testing.T) {
	t.Run("DatabaseURL takes precedence", func(t *testing.T) {
		cfg := Config{
			DatabaseURL: "postgres://custom_user:custom_pass@db.prod.internal:5432/receivables?sslmode=require",
			DBHost:      "localhost",
		}
		assert.Equal(t, "postgres://custom_user:custom_pass@db.prod.internal:5432/receivables?sslmode=require", cfg.GetDatabaseURI())
	})

	t.Run("DBURI fallback", func(t *testing.T) {
		cfg := Config{
			DBURI:  "postgres://uri_user:uri_pass@db.local:5432/db",
			DBHost: "localhost",
		}
		assert.Equal(t, "postgres://uri_user:uri_pass@db.local:5432/db", cfg.GetDatabaseURI())
	})

	t.Run("Discrete DB parameters assembled properly", func(t *testing.T) {
		cfg := Config{
			DBHost:     "pg-cluster",
			DBPort:     "5433",
			DBUser:     "admin",
			DBPassword: "secretpassword",
			DBName:     "rec_db",
			DBSSLMode:  "disable",
		}
		expected := "postgres://admin:secretpassword@pg-cluster:5433/rec_db?sslmode=disable"
		assert.Equal(t, expected, cfg.GetDatabaseURI())
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("Valid configuration with 32-byte key", func(t *testing.T) {
		cfg := Config{
			CredentialsEncryptionKey: "12345678901234567890123456789012",
			MidazOrganizationID:      "01948df2-0000-7000-8000-000000000001",
			MidazLedgerID:            "01948df2-0000-7000-8000-000000000002",
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("Valid configuration with 64-char hex key", func(t *testing.T) {
		cfg := Config{
			CredentialsEncryptionKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			MidazOrganizationID:      "01948df2-0000-7000-8000-000000000001",
			MidazLedgerID:            "01948df2-0000-7000-8000-000000000002",
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("Missing encryption key fails", func(t *testing.T) {
		cfg := Config{
			CredentialsEncryptionKey: "",
			MidazOrganizationID:      "01948df2-0000-7000-8000-000000000001",
			MidazLedgerID:            "01948df2-0000-7000-8000-000000000002",
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("Missing MidazOrganizationID fails", func(t *testing.T) {
		cfg := Config{
			CredentialsEncryptionKey: "12345678901234567890123456789012",
			MidazOrganizationID:      "",
			MidazLedgerID:            "01948df2-0000-7000-8000-000000000002",
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("Missing MidazLedgerID fails", func(t *testing.T) {
		cfg := Config{
			CredentialsEncryptionKey: "12345678901234567890123456789012",
			MidazOrganizationID:      "01948df2-0000-7000-8000-000000000001",
			MidazLedgerID:            "",
		}
		assert.Error(t, cfg.Validate())
	})
}
