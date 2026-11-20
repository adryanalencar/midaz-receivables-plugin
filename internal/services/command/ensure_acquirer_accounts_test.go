package command

import (
	"context"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnsureAcquirerAccountsHandler_Execute(t *testing.T) {
	client := new(MockMidazClient)
	client.On("EnsureClearingAccount", mock.Anything, "org_1", "led_1", domain.AcquirerSumUp).Return("@external/BRL:universal", nil)
	client.On("EnsureAcquirerInternalAccount", mock.Anything, "org_1", "led_1", domain.AcquirerSumUp).Return("@acquirer/sumup", nil)
	client.On("EnsureAcquirerInternalAccount", mock.Anything, "org_1", "led_1", domain.AcquirerPagBank).Return("@acquirer/pagbank", nil)

	handler := NewEnsureAcquirerAccountsHandler(client)
	err := handler.Execute(context.Background(), "org_1", "led_1")

	assert.NoError(t, err)
	client.AssertExpectations(t)
}
