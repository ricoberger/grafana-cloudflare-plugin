package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v6/zones"
)

func (c *client) GetZones(ctx context.Context) ([]zones.Zone, error) {
	var zs []zones.Zone

	iter := c.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})
	for iter.Next() {
		z := iter.Current()
		zs = append(zs, z)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return zs, nil
}
