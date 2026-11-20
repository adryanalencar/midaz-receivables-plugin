package command

import (
	"context"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMidazClient struct {
	mock.Mock
}

func (m *MockMidazClient) EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, acquirer)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, alias, name, accountType)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error) {
	args := m.Called(ctx, organizationID, ledgerID, alias)
	return args.Bool(0), args.Error(1)
}

func (m *MockMidazClient) PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MidazTransactionPayload) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, tx)
	return args.String(0), args.Error(1)
}

func TestEnsureAcquirerAccountsHandler_Execute(t *testing.T) {
	client := new(MockMidazClient)
	client.On("EnsureClearingAccount", mock.Anything, "org_1", "led_1", domain.AcquirerSumUp).Return("@external/acquirer_sumup", nil)
	client.On("EnsureClearingAccount", mock.Anything, "org_1", "led_1", domain.AcquirerPagBank).Return("@external/acquirer_pagbank", nil)

	handler := NewEnsureAcquirerAccountsHandler(client)
	err := handler.Execute(context.Background(), "org_1", "led_1")

	assert.NoError(t, err)
	client.AssertExpectations(t)
}
