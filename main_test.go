package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
)

const (
	testZoneNameEnvVar = "TEST_ZONE_NAME"
	manifestPath       = "testdata/selectel"
	kubeBuilderBinPath = "./_out/kubebuilder/bin"
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.
	fixture := dns.NewFixture(&selectelDNSProviderSolver{},
		dns.SetBinariesPath(kubeBuilderBinPath),
		dns.SetResolvedZone(os.Getenv(testZoneNameEnvVar)),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath(manifestPath),
		dns.SetStrict(true),
	)
	fixture.RunConformance(t)
}
