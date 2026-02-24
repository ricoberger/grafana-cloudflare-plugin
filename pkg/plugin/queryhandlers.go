package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ricoberger/grafana-cloudflare-plugin/pkg/cloudflare"
	"github.com/ricoberger/grafana-cloudflare-plugin/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (d *Datasource) handleZonesQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleZonesQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleZones, 10)
}

func (d *Datasource) handleZones(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleZones")
	defer span.End()

	var ids []string
	var names []string

	if len(d.zones) > 0 {
		for _, z := range d.zones {
			ids = append(ids, z[0])
			names = append(names, z[1])
		}
	} else {
		zones, err := d.cloudflareClient.GetZones(ctx)
		if err != nil {
			d.logger.Error("Failed to get zones", "error", err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return backend.ErrorResponseWithErrorSource(err)
		}

		for _, z := range zones {
			ids = append(ids, z.ID)
			names = append(names, z.Name)
		}
	}

	frame := data.NewFrame(
		"Zones",
		data.NewField("ids", nil, ids),
		data.NewField("names", nil, names),
	)

	frame.SetMeta(&data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
		Type:                   data.FrameTypeTable,
	})

	var response backend.DataResponse
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *Datasource) handleMetricsQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleMetricsQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleMetrics, 10)
}

func (d *Datasource) handleMetrics(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleMetrics")
	defer span.End()

	var qm models.QueryModelMetrics
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	filters := cloudflare.FiltersToGraphQL(append([]models.QueryModelMetricsFilter{{
		Field:    "datetime",
		Operator: ">=",
		Value:    query.DataQuery.TimeRange.From.Format(time.RFC3339),
	}, {
		Field:    "datetime",
		Operator: "<=",
		Value:    query.DataQuery.TimeRange.To.Format(time.RFC3339),
	}}, qm.Filters...))
	dimensions := cloudflare.DimensionsToGraphQL(qm.Dimensions)
	orderBy := cloudflare.OrderByToGraphQL(qm.OrderBy)
	if qm.Limit == 0 {
		qm.Limit = 100
	}

	d.logger.Info("handleMetrics query", "name", qm.Name, "zone", qm.Zone, "filters", filters, "dimensions", dimensions, "orderBy", orderBy, "limit", qm.Limit)
	span.SetAttributes(attribute.Key("name").String(qm.Name))
	span.SetAttributes(attribute.Key("zone").String(qm.Zone))
	span.SetAttributes(attribute.Key("filters").String(filters))
	span.SetAttributes(attribute.Key("dimensions").String(dimensions))
	span.SetAttributes(attribute.Key("orderBy").String(orderBy))
	span.SetAttributes(attribute.Key("limit").Int64(qm.Limit))

	switch {
	case qm.Name == "httpRequests":
		return d.cloudflareClient.GetHTTPRequests(ctx, qm.Zone, filters, qm.Limit)
	case strings.HasPrefix(qm.Name, "httpRequests_"):
		return d.cloudflareClient.GetHTTPRequestsAggregate(ctx, qm.Zone, qm.Name, filters, dimensions, orderBy, qm.Legend, qm.Limit, query.DataQuery.TimeRange.To)
	default:
		err := fmt.Errorf("unsupported metric name: %s", qm.Name)
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}
}
