package internal

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedDNSv2ClientRRSets struct {
	mock.Mock
	domainsV2.Client
}

var errConvertFirstArgInMock = fmt.Errorf("convert first arg")

func (client *mockedDNSv2ClientRRSets) ListRRSets(ctx context.Context, zoneID string, opts *map[string]string) (domainsV2.Listable[domainsV2.RRSet], error) {
	args := client.Called(ctx, zoneID, opts)
	rrsets, ok := args.Get(0).(domainsV2.Listable[domainsV2.RRSet])
	if !ok {
		return nil, errConvertFirstArgInMock
	}
	err := args.Error(1)

	//nolint: wrapcheck
	return rrsets, err
}

func TestGetRrsetByNameAndTypeWithoutOffset(t *testing.T) {
	t.Parallel()
	rrsetTypeForSearch := "A"
	mockedZoneID := "mocked-zone-id"
	mDNSClient := new(mockedDNSv2ClientRRSets)
	ctx := context.Background()
	opts1 := &map[string]string{
		"name":        testZoneName,
		"rrset_types": rrsetTypeForSearch,
		"limit":       "100",
		"offset":      "0",
	}
	incorrectNameForSearch := "a." + testZoneName
	rrsetWithoutOffset := domainsV2.Listable[domainsV2.RRSet](domainsV2.List[domainsV2.RRSet]{
		Count:      1,
		NextOffset: 0,
		Items: []*domainsV2.RRSet{
			{
				ID:   incorrectIDForSearch,
				Name: incorrectNameForSearch,
				Type: domainsV2.RecordType(rrsetTypeForSearch),
			},
			{
				ID:   correctIDForSearch,
				Name: testZoneName,
				Type: domainsV2.RecordType(rrsetTypeForSearch),
			},
		},
	})
	mDNSClient.On("ListRRSets", ctx, mockedZoneID, opts1).Return(rrsetWithoutOffset, nil)

	rrset, err := GetRrsetByNameAndType(ctx, mDNSClient, mockedZoneID, testZoneName, rrsetTypeForSearch)

	assert.NoError(t, err)

	assert.NotNil(t, rrset)
	assert.Equal(t, correctIDForSearch, rrset.ID)
	assert.Equal(t, testZoneName, rrset.Name)
	assert.Equal(t, rrsetTypeForSearch, string(rrset.Type))
}

func TestGetRrsetByNameAndTypeWithOffset(t *testing.T) {
	t.Parallel()
	rrsetTypeForSearch := "A"
	mockedZoneID := "mocked-zone-id"
	mDNSClient := new(mockedDNSv2ClientRRSets)
	ctx := context.Background()
	nextOffset := 3
	opts1 := &map[string]string{
		"name":        testZoneName,
		"rrset_types": rrsetTypeForSearch,
		"limit":       "100",
		"offset":      "0",
	}
	opts2 := &map[string]string{
		"name":        testZoneName,
		"rrset_types": rrsetTypeForSearch,
		"limit":       "100",
		"offset":      strconv.Itoa(nextOffset),
	}
	incorrectNameForSearch := "a." + testZoneName
	rrsetWithNextOffset := domainsV2.Listable[domainsV2.RRSet](domainsV2.List[domainsV2.RRSet]{
		Count:      1,
		NextOffset: nextOffset,
		Items: []*domainsV2.RRSet{
			{
				ID:   incorrectIDForSearch,
				Name: incorrectNameForSearch,
				Type: domainsV2.RecordType(rrsetTypeForSearch),
			},
		},
	})
	mDNSClient.On("ListRRSets", ctx, mockedZoneID, opts1).Return(rrsetWithNextOffset, nil)
	rrsetsWithoutNextOffset := domainsV2.Listable[domainsV2.RRSet](domainsV2.List[domainsV2.RRSet]{
		Count:      1,
		NextOffset: 0,
		Items: []*domainsV2.RRSet{
			{
				ID:   correctIDForSearch,
				Name: testZoneName,
				Type: domainsV2.RecordType(rrsetTypeForSearch),
			},
		},
	})
	mDNSClient.On("ListRRSets", ctx, mockedZoneID, opts2).Return(rrsetsWithoutNextOffset, nil)

	rrset, err := GetRrsetByNameAndType(ctx, mDNSClient, mockedZoneID, testZoneName, rrsetTypeForSearch)

	assert.NoError(t, err)

	assert.NotNil(t, rrset)
	assert.Equal(t, correctIDForSearch, rrset.ID)
	assert.Equal(t, testZoneName, rrset.Name)
	assert.Equal(t, rrsetTypeForSearch, string(rrset.Type))
}
