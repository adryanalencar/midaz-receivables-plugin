package providers

import (
	"fmt"
	"sync"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

var (
	ErrProviderNotRegistered = fmt.Errorf("providers: requested acquirer provider is not registered")
)

// Factory manages the lifecycle and resolution of acquirer provider adapters.
type Factory struct {
	mu        sync.RWMutex
	providers map[domain.AcquirerType]domain.AcquirerProvider
}

// NewFactory initializes an empty acquirer provider factory.
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[domain.AcquirerType]domain.AcquirerProvider),
	}
}

// Register registers a provider implementation under its AcquirerType.
func (f *Factory) Register(provider domain.AcquirerProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[provider.Type()] = provider
}

// Get resolves a registered provider adapter.
func (f *Factory) Get(acquirerType domain.AcquirerType) (domain.AcquirerProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[acquirerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotRegistered, acquirerType)
	}

	return provider, nil
}
