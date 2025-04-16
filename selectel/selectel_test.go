package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigForDNS_setupDefaultValues(t *testing.T) {
	t.Parallel()
	cfg, err := NewConfigForDNS()
	require.NoError(t, err)

	assert.Equal(t, defaultBaseURL, cfg.BaseURL)
	assert.Equal(t, defaultHTTPTimeout, cfg.HTTPTimeout)
	assert.Equal(t, minTTL, cfg.TTL)
}

func TestNewDNSProviderConfig_BadTTL(t *testing.T) {
	t.Parallel()
	testTTL := 59
	config, err := NewConfigForDNS()
	require.NoError(t, err)

	config.TTL = testTTL

	_, err = NewDNSProviderFromConfig(config)
	assert.ErrorIs(t, err, errTTLMustBeGreaterOrEqualsMinTTL)
}
