package cloudflare

import (
	"context"
	"fmt"
	"sync"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Client interface {
	GetZones(ctx context.Context) ([]zones.Zone, error)
	GetHTTPRequests(ctx context.Context, zoneId, filters string, limit int64) backend.DataResponse
	GetHTTPRequestsVolumes(ctx context.Context, zoneId, filtersInfo, filtersWarning, filtersError, filtersCritical string) backend.DataResponse
	GetHTTPRequestsAggregate(ctx context.Context, zoneId, metricName, aggregation, filters, dimensions, orderBy, legend string, limit int64, timeTo time.Time) backend.DataResponse
	GetFirewallEvents(ctx context.Context, zoneId, filters string, limit int64) backend.DataResponse
	GetFirewallEventsVolumes(ctx context.Context, zoneId, filtersInfo, filtersWarning, filtersError, filtersCritical string) backend.DataResponse
	GetFirewallEventsAggregate(ctx context.Context, zoneId, filters, dimensions, orderBy, legend string, limit int64, timeTo time.Time) backend.DataResponse
}

type client struct {
	client        *cloudflare.Client
	logger        log.Logger
	zoneCache     []zones.Zone
	zoneCacheTime time.Time
	zoneCacheLock sync.RWMutex
}

func NewClient(logger log.Logger, authMethod, apiToken, apiEmail, apiKey string) (Client, error) {
	opts := []option.RequestOption{option.WithEnvironmentProduction()}
	switch authMethod {
	case "apiToken":
		opts = append(opts, option.WithAPIToken(apiToken))
	case "apiKey":
		opts = append(opts, option.WithAPIEmail(apiEmail))
		opts = append(opts, option.WithAPIKey(apiKey))
	default:
		return nil, fmt.Errorf("invalid authentication method")
	}

	cloudflareClient := cloudflare.NewClient(opts...)

	return &client{
		client: cloudflareClient,
		logger: logger,
	}, nil
}
