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

var ErrRrsetNotFound = errors.New("rrset not found")

func GetRrsetByNameAndType(ctx context.Context, client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], zoneID, rrsetName, rrsetType string) (*domainsV2.RRSet, error) {
	rrsetNameUnicode, err := idna.ToUnicode(rrsetName)
	if err != nil {
		return nil, fmt.Errorf("convert rrset name to unicode: %w", err)
	}

	optsForSearchRrset := map[string]string{
		"name":        rrsetNameUnicode,
		"rrset_types": rrsetType,
		"limit":       "100",
		"offset":      "0",
	}

	regexpRRSetWithDotOrNot, err := regexp.Compile(fmt.Sprintf("^%s.?", rrsetNameUnicode))
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
