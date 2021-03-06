// Package selectel implements a DNS provider for solving the DNS-01 challenge using Selectel Domains API.
// Selectel Domain API reference: https://kb.selectel.com/23136054.html
// Token: https://my.selectel.ru/profile/apikeys
package selectel

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/selectel/cert-manager-webhook-selectel/selectel/internal"
)

const (
	defaultBaseURL            = "https://api.selectel.ru/domains/v1"
	minTTL                    = 60
	defaultPropagationTimeout = 120
	defaultPollingInterval    = 2
	defaultHTTPTimeout        = 30
)

const (
	envNamespace             = "SELECTEL_"
	baseURLEnvVar            = envNamespace + "BASE_URL"
	apiTokenEnvVar           = envNamespace + "API_TOKEN"
	ttlEnvVar                = envNamespace + "TTL"
	propagationTimeoutEnvVar = envNamespace + "PROPAGATION_TIMEOUT"
	pollingIntervalEnvVar    = envNamespace + "POLLING_INTERVAL"
	httpTimeoutEnvVar        = envNamespace + "HTTP_TIMEOUT"
)

const (
	errConfigAbsent       = "the configuration of the DNS provider is absent"
	errCredentialsMissing = "credentials missing"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            getEnvOrDefaultString(baseURLEnvVar, defaultBaseURL),
		TTL:                getEnvOrDefaultInt(ttlEnvVar, minTTL),
		PropagationTimeout: getEnvOrDefaultSecond(propagationTimeoutEnvVar, defaultPropagationTimeout),
		PollingInterval:    getEnvOrDefaultSecond(pollingIntervalEnvVar, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: getEnvOrDefaultSecond(httpTimeoutEnvVar, defaultHTTPTimeout),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Selectel Domains API.
// API token must be passed in the environment variable SELECTEL_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	token, err := getEnvOrErrorString(apiTokenEnvVar)
	if err != nil {
		return nil, fmt.Errorf("selectel: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = token

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for selectel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("selectel: %s", errConfigAbsent)
	}

	if config.Token == "" {
		return nil, fmt.Errorf("selectel: %s", errCredentialsMissing)
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("selectel: invalid TTL, TTL (%d) must be greater than %d",
			config.TTL,
			minTTL)
	}

	return &DNSProvider{
		config: config,
		client: internal.NewClient(internal.ClientOpts{
			BaseURL:    config.BaseURL,
			Token:      config.Token,
			HTTPClient: config.HTTPClient,
		}),
	}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(fqdn, value string) error {
	domainObj, err := d.client.GetDomainByName(fqdn)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	txtRecord := internal.Record{
		Type:    "TXT",
		TTL:     d.config.TTL,
		Name:    fqdn,
		Content: value,
	}
	_, err = d.client.AddRecord(domainObj.ID, txtRecord)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(fqdn, key string) error {
	domainObj, err := d.client.GetDomainByName(fqdn)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	records, err := d.client.ListRecords(domainObj.ID)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	// Delete records with specific FQDN and key
	var lastErr error
	for _, record := range records {
		if record.Name == unFQDN(fqdn) && record.Content == key {
			err = d.client.DeleteRecord(domainObj.ID, record.ID)
			if err != nil {
				lastErr = fmt.Errorf("selectel: %v", err)
			}
		}
	}

	return lastErr
}

func unFQDN(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}

func getEnvOrDefaultString(envVar, defaultValue string) string {
	v := os.Getenv(envVar)
	if v == "" {
		return defaultValue
	}
	return v
}

func getEnvOrErrorString(envVar string) (string, error) {
	v := getEnvOrDefaultString(envVar, "")
	if v == "" {
		return "", fmt.Errorf("%s is missing", envVar)
	}
	return v, nil
}

func getEnvOrDefaultInt(envVar string, defaultValue int) int {
	v, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvOrDefaultSecond(envVar string, defaultValue int) time.Duration {
	return time.Duration(getEnvOrDefaultInt(envVar, defaultValue)) * time.Second
}
