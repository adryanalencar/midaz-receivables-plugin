package command

import (
	"context"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateReceivablesAccountHandler_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)

		midazClient.On("CreateAccount", mock.Anything, "org_1", "led_1", "@merchant_a/receivables", "Merchant Alpha", "creditCard").
			Return("@merchant_a/receivables", nil)

		repo.On("CreateReceivablesAccount", mock.Anything, mock.MatchedBy(func(acc *domain.ReceivablesAccount) bool {
			return acc.AccountAlias == "@merchant_a/receivables" &&
				acc.Name == "Merchant Alpha" &&
				acc.CheckingAccountAlias == "@merchant_a/checking"
		})).Return(nil)

		handler := NewCreateReceivablesAccountHandler(repo, midazClient)

		acc, err := handler.Execute(context.Background(), CreateReceivablesAccountInput{
			OrganizationID:              "org_1",
			LedgerID:                    "led_1",
			Alias:                       "merchant_a",
			Name:                        "Merchant Alpha",
			CNPJ:                        "12345678000199",
			DefaultCheckingAccountAlias: "merchant_a/checking",
		})

		require.NoError(t, err)
		assert.Equal(t, "@merchant_a/receivables", acc.AccountAlias)
		assert.Equal(t, "@merchant_a/checking", acc.CheckingAccountAlias)
		midazClient.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		handler := NewCreateReceivablesAccountHandler(nil, nil)

		_, err := handler.Execute(context.Background(), CreateReceivablesAccountInput{})
		assert.ErrorIs(t, err, ErrEmptyAccountAlias)

		_, err = handler.Execute(context.Background(), CreateReceivablesAccountInput{Alias: "valid"})
		assert.ErrorIs(t, err, ErrEmptyAccountName)
	})
}
