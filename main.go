package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/selectel/cert-manager-webhook-selectel/selectel"
	extAPI "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	providerName = "selectel"

	groupNameEnvVar = "GROUP_NAME"
)

func main() {
	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		panic(fmt.Sprintf("%s must be specified", groupNameEnvVar))
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided groupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(groupName,
		&selectelDNSProviderSolver{},
	)
}

// selectelDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type selectelDNSProviderSolver struct {
	client *kubernetes.Clientset
}

// selectelDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
type selectelDNSProviderConfig struct {
	APIKeySecretRef cmmeta.SecretKeySelector `json:"apiKeySecretRef"`

	// +optional
	TTL int `json:"ttl"`

	// +optional
	Timeout int `json:"timeout"`

	// +optional
	PropagationTimeout int `json:"propagationTimeout"`

	// +optional
	PollingInterval int `json:"pollingInterval"`
}

func (c *selectelDNSProviderSolver) validate(cfg *selectelDNSProviderConfig) error {
	// Try to load the API key
	if cfg.APIKeySecretRef.LocalObjectReference.Name == "" {
		return errors.New("API token field were not provided")
	}

	return nil
}

func (c *selectelDNSProviderSolver) provider(cfg *selectelDNSProviderConfig,
	namespace string) (*selectel.DNSProvider, error) {

	if err := c.validate(cfg); err != nil {
		return nil, err
	}

	sec, err := c.client.CoreV1().
		Secrets(namespace).
		Get(cfg.APIKeySecretRef.LocalObjectReference.Name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secBytes, ok := sec.Data[cfg.APIKeySecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret \"%s/%s\"",
			cfg.APIKeySecretRef.Key,
			cfg.APIKeySecretRef.LocalObjectReference.Name,
			namespace)
	}

	// Create provider
	providerConfig := selectel.NewDefaultConfig()

	providerConfig.Token = string(secBytes)

	if cfg.PropagationTimeout > 0 {
		providerConfig.PropagationTimeout = time.Duration(cfg.PropagationTimeout) * time.Second
	}

	if cfg.PollingInterval > 0 {
		providerConfig.PollingInterval = time.Duration(cfg.PollingInterval) * time.Second
	}

	if cfg.TTL > 0 {
		providerConfig.TTL = cfg.TTL
	}

	if cfg.Timeout > 0 {
		providerConfig.HTTPClient.Timeout = time.Duration(cfg.Timeout) * time.Second
	}

	return selectel.NewDNSProviderConfig(providerConfig)
}

// Return DNS provider name.
func (c *selectelDNSProviderSolver) Name() string {
	return providerName
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *selectelDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	// Load environment variables and create new Selectel DNS provider
	provider, err := c.provider(&cfg, ch.ResourceNamespace)
	if err != nil {
		return err
	}

	return provider.Present(ch.ResolvedFQDN, ch.Key)
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *selectelDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	// Load environment variables and create new Selectel DNS provider
	provider, err := c.provider(&cfg, ch.ResourceNamespace)

	if err != nil {
		return err
	}

	// Remove TXT DNS record
	return provider.CleanUp(ch.ResolvedFQDN, ch.Key)
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *selectelDNSProviderSolver) Initialize(kubeClientCfg *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientCfg)
	if err != nil {
		return err
	}

	c.client = cl
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extAPI.JSON) (selectelDNSProviderConfig, error) {
	cfg := selectelDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
