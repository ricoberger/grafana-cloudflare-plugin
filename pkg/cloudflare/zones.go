package cloudflare

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go/v6/zones"
)

func (c *client) GetZones(ctx context.Context) ([]zones.Zone, error) {
	if c.zoneCacheTime.Add(60 * time.Minute).After(time.Now()) {
		c.zoneCacheLock.RLock()
		defer c.zoneCacheLock.RUnlock()
		return c.zoneCache, nil
	}

	var zs []zones.Zone

	iter := c.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})
	for iter.Next() {
		z := iter.Current()
		zs = append(zs, z)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	c.zoneCacheLock.Lock()
	defer c.zoneCacheLock.Unlock()
	c.zoneCache = zs
	c.zoneCacheTime = time.Now()

	return zs, nil
}
