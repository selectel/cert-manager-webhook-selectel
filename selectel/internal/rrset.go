package internal

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

var ErrRrsetNotFound = fmt.Errorf("rrset not found")

func GetRrsetByNameAndType(ctx context.Context, client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], zoneID, rrsetName, rrsetType string) (*domainsV2.RRSet, error) {
	optsForSearchRrset := map[string]string{
		"name":        rrsetName,
		"rrset_types": rrsetType,
		"limit":       "100",
		"offset":      "0",
	}

	regexpRRSetWithDotOrNot, err := regexp.Compile(fmt.Sprintf("^%s.?", rrsetName))
	if err != nil {
		return nil, fmt.Errorf("compile regexp: %w", err)
	}

	for {
		rrsets, err := client.ListRRSets(ctx, zoneID, &optsForSearchRrset)
		if err != nil {
			return nil, fmt.Errorf("list rrsets: %w", err)
		}
		for _, rrset := range rrsets.GetItems() {
			if regexpRRSetWithDotOrNot.MatchString(rrset.Name) && string(rrset.Type) == rrsetType {
				return rrset, nil
			}
		}
		optsForSearchRrset["offset"] = strconv.Itoa(rrsets.GetNextOffset())
		if rrsets.GetNextOffset() == 0 {
			break
		}
	}

	return nil, ErrRrsetNotFound
}
