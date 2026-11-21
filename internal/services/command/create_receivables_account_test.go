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
	t.Run("Success with Subaccounts and Parent Account Hierarchy", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)
		matcherClient := new(MockMatcherClient)

		parentID := "01948df2-mother-0001"
		receivablesUUID := "01948df2-receivables-0002"

		// Expect creation of Receivables Subaccount
		midazClient.On("CreateAccount", mock.Anything, "org_1", "led_1", "@merchant_a/receivables", "Merchant Alpha", "creditCard", &parentID).
			Return(receivablesUUID, nil)

		// Expect creation of MDR Revenue Subaccount
		midazClient.On("CreateAccount", mock.Anything, "org_1", "led_1", "@merchant_a/receivables/mdr_revenue", "Merchant Alpha - MDR Revenue", "revenue", &receivablesUUID).
			Return("01948df2-mdr-0003", nil)

		// Expect creation of Fee/Tariff Subaccount
		midazClient.On("CreateAccount", mock.Anything, "org_1", "led_1", "@merchant_a/receivables/fees", "Merchant Alpha - Tariffs & Deductions", "deposit", &receivablesUUID).
			Return("01948df2-fees-0004", nil)

		// Expect Matcher Context Provisioning
		matcherClient.On("EnsureContextAndSources", mock.Anything).
			Return("ctx_merchant_a", "src_a", "src_b", nil)

		repo.On("CreateReceivablesAccount", mock.Anything, mock.MatchedBy(func(acc *domain.ReceivablesAccount) bool {
			return acc.AccountAlias == "@merchant_a/receivables" &&
				acc.MDRRevenueAccountAlias == "@merchant_a/receivables/mdr_revenue" &&
				acc.FeeAccountAlias == "@merchant_a/receivables/fees" &&
				acc.MotherAccountAlias == "@merchant_a" &&
				acc.ParentAccountID != nil && *acc.ParentAccountID == parentID &&
				acc.MatcherContextID == "ctx_merchant_a" &&
				acc.Name == "Merchant Alpha" &&
				acc.CheckingAccountAlias == "@merchant_a/checking" &&
				acc.ExternalAccountAlias == "@external/BRL:custom"
		})).Return(nil)

		handler := NewCreateReceivablesAccountHandler(repo, midazClient, matcherClient)

		acc, err := handler.Execute(context.Background(), CreateReceivablesAccountInput{
			OrganizationID:              "org_1",
			LedgerID:                    "led_1",
			ParentAccountID:             &parentID,
			MotherAccountAlias:          "@merchant_a",
			Alias:                       "merchant_a",
			Name:                        "Merchant Alpha",
			CNPJ:                        "12345678000199",
			DefaultCheckingAccountAlias: "merchant_a/checking",
			ExternalAccountAlias:        "@external/BRL:custom",
		})

		require.NoError(t, err)
		assert.Equal(t, "@merchant_a/receivables", acc.AccountAlias)
		assert.Equal(t, "@merchant_a/receivables/mdr_revenue", acc.MDRRevenueAccountAlias)
		assert.Equal(t, "@merchant_a/receivables/fees", acc.FeeAccountAlias)
		assert.Equal(t, "@merchant_a/checking", acc.CheckingAccountAlias)
		assert.Equal(t, "@external/BRL:custom", acc.ExternalAccountAlias)
		assert.Equal(t, "ctx_merchant_a", acc.MatcherContextID)
		midazClient.AssertExpectations(t)
		matcherClient.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		handler := NewCreateReceivablesAccountHandler(nil, nil, nil)

		_, err := handler.Execute(context.Background(), CreateReceivablesAccountInput{})
		assert.ErrorIs(t, err, ErrEmptyAccountAlias)

		_, err = handler.Execute(context.Background(), CreateReceivablesAccountInput{Alias: "valid"})
		assert.ErrorIs(t, err, ErrEmptyAccountName)
	})
}
