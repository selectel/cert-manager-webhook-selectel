package internal

import (
	"context"
	"strconv"
	"testing"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedDNSv2ClientZones struct {
	mock.Mock
	domainsV2.Client
}

func (client *mockedDNSv2ClientZones) ListZones(ctx context.Context, opts *map[string]string) (domainsV2.Listable[domainsV2.Zone], error) {
	args := client.Called(ctx, opts)
	zones, ok := args.Get(0).(domainsV2.Listable[domainsV2.Zone])
	if !ok {
		return nil, errConvertFirstArgInMock
	}
	err := args.Error(1)

	//nolint: wrapcheck
	return zones, err
}

const (
	testZoneName         = "test.xyz."
	correctIDForSearch   = "mocked-uuid-2"
	incorrectIDForSearch = "mocked-uuid-1"
)

func TestGetZoneByNameWithoutOffset(t *testing.T) {
	t.Parallel()
	mDNSClient := new(mockedDNSv2ClientZones)
	ctx := context.Background()
	opts1 := &map[string]string{
		"filter": testZoneName,
		"limit":  "100",
		"offset": "0",
	}
	incorrectNameForSearch := "a." + testZoneName
	zonesWithoutOffset := domainsV2.Listable[domainsV2.Zone](domainsV2.List[domainsV2.Zone]{
		Count:      1,
		NextOffset: 0,
		Items: []*domainsV2.Zone{
			{
				ID:   incorrectIDForSearch,
				Name: incorrectNameForSearch,
			},
			{
				ID:   correctIDForSearch,
				Name: testZoneName,
			},
		},
	})
	mDNSClient.On("ListZones", ctx, opts1).Return(zonesWithoutOffset, nil)

	zone, err := GetZoneByName(ctx, mDNSClient, testZoneName)
	assert.NoError(t, err)

	assert.NotNil(t, zone)
	assert.Equal(t, correctIDForSearch, zone.ID)
	assert.Equal(t, testZoneName, zone.Name)
}

func TestGetZoneByNameWithOffset(t *testing.T) {
	t.Parallel()
	mDNSClient := new(mockedDNSv2ClientZones)
	ctx := context.Background()
	nextOffset := 3
	opts1 := &map[string]string{
		"filter": testZoneName,
		"limit":  "100",
		"offset": "0",
	}
	opts2 := &map[string]string{
		"filter": testZoneName,
		"limit":  "100",
		"offset": strconv.Itoa(nextOffset),
	}
	incorrectNameForSearch := "a." + testZoneName
	zonesWithNextOffset := domainsV2.Listable[domainsV2.Zone](domainsV2.List[domainsV2.Zone]{
		Count:      1,
		NextOffset: nextOffset,
		Items: []*domainsV2.Zone{
			{
				ID:   incorrectIDForSearch,
				Name: incorrectNameForSearch,
			},
		},
	})
	mDNSClient.On("ListZones", ctx, opts1).Return(zonesWithNextOffset, nil)
	zonesWithoutNextOffset := domainsV2.Listable[domainsV2.Zone](domainsV2.List[domainsV2.Zone]{
		Count:      1,
		NextOffset: 0,
		Items: []*domainsV2.Zone{
			{
				ID:   correctIDForSearch,
				Name: testZoneName,
			},
		},
	})
	mDNSClient.On("ListZones", ctx, opts2).Return(zonesWithoutNextOffset, nil)

	zone, err := GetZoneByName(ctx, mDNSClient, testZoneName)
	assert.NoError(t, err)

	assert.NotNil(t, zone)
	assert.Equal(t, correctIDForSearch, zone.ID)
	assert.Equal(t, testZoneName, zone.Name)
}
