package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
	"github.com/stretchr/testify/require"
)

// The fake widens the read-modify-write window on purpose: Present reads the RRset
// through ListRRSets and writes it back through UpdateRRSet, so a slow read followed
// by a slow write is exactly what a real API round-trip looks like.
const (
	fakeReadDelay  = 20 * time.Millisecond
	fakeWriteDelay = 5 * time.Millisecond
)

var errFakeBadType = errors.New("fake: unexpected rrset type")

type fakeDNS struct {
	mu      sync.Mutex
	zone    domainsV2.Zone
	rrsets  map[string]*domainsV2.RRSet
	nextID  int
	creates int
}

func newFakeDNS() *fakeDNS {
	return &fakeDNS{
		zone:   domainsV2.Zone{ID: "zone-id", Name: "example.com."},
		rrsets: make(map[string]*domainsV2.RRSet),
	}
}

func (f *fakeDNS) GetZone(_ context.Context, _ string, _ *map[string]string) (*domainsV2.Zone, error) {
	return &f.zone, nil
}

func (f *fakeDNS) ListZones(_ context.Context, _ *map[string]string) (domainsV2.Listable[domainsV2.Zone], error) {
	return domainsV2.List[domainsV2.Zone]{
		Count: 1,
		Items: []*domainsV2.Zone{&f.zone},
	}, nil
}

func (f *fakeDNS) CreateZone(_ context.Context, _ domainsV2.Creatable) (*domainsV2.Zone, error) {
	return &f.zone, nil
}

func (f *fakeDNS) DeleteZone(_ context.Context, _ string) error { return nil }

func (f *fakeDNS) UpdateZoneState(_ context.Context, _ string, _ bool) error { return nil }

func (f *fakeDNS) UpdateZoneComment(_ context.Context, _ string, _ string) error { return nil }

func (f *fakeDNS) CreateRRSet(_ context.Context, _ string, creatable domainsV2.Creatable) (*domainsV2.RRSet, error) {
	time.Sleep(fakeWriteDelay)

	rrset, ok := creatable.(*domainsV2.RRSet)
	if !ok {
		return nil, errFakeBadType
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	key := rrset.Name + "|" + string(rrset.Type)
	if _, exists := f.rrsets[key]; exists {
		// Same error the API returns on a duplicate create.
		return nil, errors.New("409 this_rrset_is_already_exists")
	}
	f.nextID++
	f.creates++
	stored := *rrset
	stored.ID = fmt.Sprintf("rrset-%d", f.nextID)
	f.rrsets[key] = &stored

	out := stored

	return &out, nil
}

func (f *fakeDNS) GetRRSet(_ context.Context, _, rrsetID string) (*domainsV2.RRSet, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, rrset := range f.rrsets {
		if rrset.ID == rrsetID {
			out := *rrset

			return &out, nil
		}
	}

	return nil, errors.New("404 rrset_not_found")
}

func (f *fakeDNS) ListRRSets(_ context.Context, _ string, _ *map[string]string) (domainsV2.Listable[domainsV2.RRSet], error) {
	time.Sleep(fakeReadDelay)

	f.mu.Lock()
	defer f.mu.Unlock()

	items := make([]*domainsV2.RRSet, 0, len(f.rrsets))
	for _, rrset := range f.rrsets {
		out := *rrset
		items = append(items, &out)
	}

	return domainsV2.List[domainsV2.RRSet]{Count: len(items), Items: items}, nil
}

func (f *fakeDNS) UpdateRRSet(_ context.Context, _, rrsetID string, updatable domainsV2.Updatable) error {
	time.Sleep(fakeWriteDelay)

	rrset, ok := updatable.(*domainsV2.RRSet)
	if !ok {
		return errFakeBadType
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for _, stored := range f.rrsets {
		if stored.ID == rrsetID {
			stored.Records = rrset.Records
			stored.TTL = rrset.TTL

			return nil
		}
	}

	return errors.New("404 rrset_not_found")
}

func (f *fakeDNS) DeleteRRSet(_ context.Context, _, rrsetID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for key, rrset := range f.rrsets {
		if rrset.ID == rrsetID {
			delete(f.rrsets, key)

			return nil
		}
	}

	return errors.New("404 rrset_not_found")
}

func (f *fakeDNS) WithHeaders(_ http.Header) domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet] {
	return f
}

// contents returns the TXT values currently published for name.
func (f *fakeDNS) contents(name string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	for key, rrset := range f.rrsets {
		if key == name+"|"+string(domainsV2.TXT) {
			out := make([]string, 0, len(rrset.Records))
			for _, record := range rrset.Records {
				out = append(out, record.Content)
			}

			return out
		}
	}

	return nil
}

func newTestProvider(client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet]) *DNSProvider {
	return &DNSProvider{
		config:    &Config{TTL: minTTL},
		dnsClient: client,
	}
}

// A Certificate with both example.com and *.example.com puts two challenges on the
// same name. cert-manager runs them concurrently (--max-concurrent-challenges = 60 by
// default), so both values must survive.
func TestPresentKeepsBothValuesForSameName(t *testing.T) {
	t.Parallel()

	const (
		zone = "example.com"
		name = "_acme-challenge.example.com"
	)

	client := newFakeDNS()
	provider := newTestProvider(client)

	values := []string{"value-one", "value-two"}
	errs := make([]error, len(values))

	var wg sync.WaitGroup

	for i, value := range values {
		wg.Add(1)

		go func() {
			defer wg.Done()

			errs[i] = provider.Present(zone, name, value)
		}()
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "Present #%d must not fail", i)
	}

	published := client.contents(name)
	require.Len(t, published, len(values), "both TXT values must be published, got %v", published)

	for _, value := range values {
		require.Contains(t, published, fmt.Sprintf("%q", value))
	}
}

// CleanUp of one challenge must not drop the other challenge's value.
func TestCleanUpKeepsOtherValue(t *testing.T) {
	t.Parallel()

	const (
		zone = "example.com"
		name = "_acme-challenge.example.com"
	)

	client := newFakeDNS()
	provider := newTestProvider(client)

	require.NoError(t, provider.Present(zone, name, "value-one"))
	require.NoError(t, provider.Present(zone, name, "value-two"))

	require.NoError(t, provider.CleanUp(zone, name, "value-one"))

	published := client.contents(name)
	require.Len(t, published, 1, "one value must be left, got %v", published)
	require.Contains(t, published, "\"value-two\"")
}

// The same value presented twice is a no-op, not a duplicate record.
func TestPresentIsIdempotent(t *testing.T) {
	t.Parallel()

	const (
		zone = "example.com"
		name = "_acme-challenge.example.com"
	)

	client := newFakeDNS()
	provider := newTestProvider(client)

	require.NoError(t, provider.Present(zone, name, "value-one"))
	require.NoError(t, provider.Present(zone, name, "value-one"))

	require.Len(t, client.contents(name), 1)
	require.Equal(t, 1, client.creates, "the RRset must be created once")
}

// The update path is where the lost update lives: both challenges read the same
// existing RRset (it already carries one value), append their own value and write the
// full list back. Without serialisation the second write drops the first value.
func TestPresentKeepsBothValuesOnExistingRrset(t *testing.T) {
	t.Parallel()

	const (
		zone = "example.com"
		name = "_acme-challenge.example.com"
	)

	client := newFakeDNS()
	provider := newTestProvider(client)

	// Seed: the RRset exists and already holds a value from an earlier challenge.
	require.NoError(t, provider.Present(zone, name, "value-seed"))

	values := []string{"value-one", "value-two"}
	errs := make([]error, len(values))

	start := make(chan struct{})

	var wg sync.WaitGroup

	for i, value := range values {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			errs[i] = provider.Present(zone, name, value)
		}()
	}

	close(start)
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "Present #%d must not fail", i)
	}

	published := client.contents(name)
	require.Len(t, published, 3, "seed and both challenge values must survive, got %v", published)

	for _, value := range append(values, "value-seed") {
		require.Contains(t, published, fmt.Sprintf("%q", value))
	}
}
