package internal

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
	"golang.org/x/net/idna"
)

var ErrZoneNotFound = errors.New("zone not found")

func GetZoneByName(ctx context.Context, client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], zoneName string) (*domainsV2.Zone, error) {
	zoneNameUnicode, err := idna.ToUnicode(zoneName)
	if err != nil {
		return nil, fmt.Errorf("convert zone name to unicode: %w", err)
	}

	optsForSearchZone := map[string]string{
		"filter": zoneNameUnicode,
		"limit":  "100",
		"offset": "0",
	}

	regexpZoneWithDotOrNot, err := regexp.Compile(fmt.Sprintf("^%s.?", zoneNameUnicode))
	if err != nil {
		return nil, fmt.Errorf("compile regexp: %w", err)
	}

	for {
		zones, err := client.ListZones(ctx, &optsForSearchZone)
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}

		for _, zone := range zones.GetItems() {
			if regexpZoneWithDotOrNot.MatchString(zone.Name) {
				return zone, nil
			}
		}

		optsForSearchZone["offset"] = strconv.Itoa(zones.GetNextOffset())
		if zones.GetNextOffset() == 0 {
			break
		}
	}

	return nil, ErrZoneNotFound
}
