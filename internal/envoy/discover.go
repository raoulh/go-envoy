package envoy

import (
	"context"
	"time"

	"github.com/brutella/dnssd"
)

func Discover() (string, error) {
	discovered := "envoy"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	found := func(e dnssd.BrowseEntry) {
		// look through the list of IPs, pick something IPv4
		for _, ipa := range e.IPs {
			if ipa.To4() != nil {
				discovered = ipa.String()
				cancel()
				return
			}
		}
	}

	if err := dnssd.LookupType(ctx, "_enphase-envoy._tcp.local.", found, reject); err != nil {
		if err.Error() != "context canceled" {
			logging.Debugf("discovery: %v\n", err)
			return "", err
		}
	}
	return discovered, nil
}

func reject(e dnssd.BrowseEntry) {
	logging.Debugf("dnssd-lookup: %+v", e)
}
