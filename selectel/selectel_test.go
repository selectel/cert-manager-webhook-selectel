package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigForDNS_setupDefaultValues(t *testing.T) {
	t.Parallel()
	cfg, err := NewConfigForDNS()
	assert.NoError(t, err)

	assert.Equal(t, cfg.BaseURL, defaultBaseURL)
	assert.Equal(t, cfg.HTTPTimeout, defaultHTTPTimeout)
	assert.Equal(t, cfg.TTL, minTTL)
}

func TestNewDNSProviderConfig_BadTTL(t *testing.T) {
	t.Parallel()
	testTTL := 59
	config, err := NewConfigForDNS()
	assert.NoError(t, err)
	config.TTL = testTTL

	_, err = NewDNSProviderFromConfig(config)
	assert.ErrorIs(t, err, errTTLMustBeGreaterOrEqualsMinTTL)
}
