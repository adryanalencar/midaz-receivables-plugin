package providers

import (
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/aditum"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/pagbank"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/sumup"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory_RegisterAndGet(t *testing.T) {
	factory := NewFactory()

	sumupAdapter := sumup.NewAdapter("", nil)
	pagbankAdapter := pagbank.NewAdapter("", nil)
	aditumAdapter := aditum.NewAdapter("", "", nil)

	factory.Register(sumupAdapter)
	factory.Register(pagbankAdapter)
	factory.Register(aditumAdapter)

	// Resolve SumUp
	p1, err := factory.Get(domain.AcquirerSumUp)
	require.NoError(t, err)
	assert.Equal(t, domain.AcquirerSumUp, p1.Type())

	// Resolve PagBank
	p2, err := factory.Get(domain.AcquirerPagBank)
	require.NoError(t, err)
	assert.Equal(t, domain.AcquirerPagBank, p2.Type())

	// Resolve Aditum
	pAditum, err := factory.Get(domain.AcquirerAditum)
	require.NoError(t, err)
	assert.Equal(t, domain.AcquirerAditum, pAditum.Type())

	// Resolve unknown
	p3, err := factory.Get("UNKNOWN_ACQUIRER")
	assert.ErrorIs(t, err, ErrProviderNotRegistered)
	assert.Nil(t, p3)
}
