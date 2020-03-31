package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDNSProvider(t *testing.T) {
	testToken := "123"

	assert.NoError(t, os.Setenv(apiTokenEnvVar, testToken))
	defer os.Unsetenv(apiTokenEnvVar)

	p, err := NewDNSProvider()
	assert.NoError(t, err)
	assert.NotNil(t, p.config)
	assert.NotNil(t, p.client)
	assert.Equal(t, p.config.Token, testToken)
}

func TestNewDNSProvider_NoAPIKey(t *testing.T) {
	assert.NoError(t, os.Setenv(apiTokenEnvVar, ""))
	defer os.Unsetenv(apiTokenEnvVar)

	_, err := NewDNSProvider()
	if assert.Error(t, err) {
		assert.EqualError(t, err, fmt.Sprintf("selectel: %s is missing", apiTokenEnvVar))
	}
}

func TestNewDNSProviderConfig_NoConfig(t *testing.T) {
	_, err := NewDNSProviderConfig(nil)
	if assert.Error(t, err) {
		assert.EqualError(t, err, fmt.Sprintf("selectel: %s", errConfigAbsent))
	}
}

func TestNewDNSProviderConfig_BadTTL(t *testing.T) {
	testTTL := 59
	testAPIKey := "123"
	config := NewDefaultConfig()
	config.TTL = testTTL
	config.Token = testAPIKey

	_, err := NewDNSProviderConfig(config)
	if assert.Error(t, err) {
		assert.EqualError(t, err,
			fmt.Sprintf("selectel: invalid TTL, TTL (%d) must be greater than %d", testTTL, minTTL))
	}
}
