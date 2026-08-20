package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ricoberger/grafana-cloudflare-plugin/pkg/models"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/shared"
	"github.com/cloudflare/cloudflare-go/v6/workers"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// workersLogsDatasets are the names of the Workers Observability datasets
// which contain the logs of Cloudflare Workers. Besides the standard
// "cloudflare-workers" dataset, the "containers" and "otel" datasets are
// included so that logs from Workers for Platform and container workloads
// are returned as well, matching the datasets used by the Cloudflare
// observability dashboard.
var workersLogsDatasets = []string{"containers", "otel", "cloudflare-workers"}

// workersLogsQueryID is the id used for all queries against the Workers
// Observability API. Since all queries are executed as dry run, the queries
// are not saved and the id is never persisted.
const workersLogsQueryID = "grafana-cloudflare-plugin"

var workersLogsFilterOperations = map[string]workers.ObservabilityTelemetryQueryParamsParametersFiltersOperation{
	"=":  workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationEquals,
	"!=": workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationNotEquals,
	">":  workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationGreater,
	"<":  workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationLess,
	">=": workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationGreaterOrEquals,
	"<=": workers.ObservabilityTelemetryQueryParamsParametersFiltersOperationLessOrEquals,
}

// workersLogsFiltersToTelemetry converts the filters from the query model to
// the filters format used by the Workers Observability API. Filters where the
// field is "-" are ignored, because this is used to indicate that the filter
// should be ignored (see FiltersToGraphQL).
func workersLogsFiltersToTelemetry(filters []models.QueryModelMetricsFilter) []workers.ObservabilityTelemetryQueryParamsParametersFilterUnion {
	var telemetryFilters []workers.ObservabilityTelemetryQueryParamsParametersFilterUnion

	for _, f := range filters {
		if f.Field == "-" {
			continue
		}

		operation, ok := workersLogsFilterOperations[f.Operator]
		if !ok {
			continue
		}

		filter := workers.ObservabilityTelemetryQueryParamsParametersFilter{
			Key:       cloudflare.F(f.Field),
			Kind:      cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFiltersKindFilter),
			Operation: cloudflare.F(operation),
		}

		if v, err := strconv.ParseFloat(f.Value, 64); err == nil {
			filter.Type = cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFiltersTypeNumber)
			filter.Value = cloudflare.F[any](v)
		} else {
			filter.Type = cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFiltersTypeString)
			filter.Value = cloudflare.F[any](f.Value)
		}

		telemetryFilters = append(telemetryFilters, filter)
	}

	return telemetryFilters
}

// workersLogsLevelToSeverity converts the level from the "$metadata.level"
// field of a Workers Observability event to a severity which is known by
// Grafana.
func workersLogsLevelToSeverity(level string) string {
	switch level {
	case "fatal":
		return "critical"
	case "error":
		return "error"
	case "warn", "warning":
		return "warning"
	case "debug":
		return "debug"
	default:
		return "info"
	}
}

// workersLogsUnionToString converts a union value returned by the Workers
// Observability API to it's string representation.
func workersLogsUnionToString(value any) string {
	switch v := value.(type) {
	case shared.UnionString:
		return string(v)
	case shared.UnionFloat:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case shared.UnionInt:
		return strconv.FormatInt(int64(v), 10)
	case shared.UnionBool:
		return strconv.FormatBool(bool(v))
	default:
		return fmt.Sprintf("%v", value)
	}
}

// workersLogsParseSeriesTime parses the time of a series returned by the
// Workers Observability API, which can either be a datetime formatted string
// (e.g. "2026-08-05 12:09:00", in UTC), a RFC3339 formatted string or a Unix
// timestamp in milliseconds.
func workersLogsParseSeriesTime(value string) (time.Time, error) {
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if ms, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.UnixMilli(ms), nil
	}
	return time.Time{}, fmt.Errorf("failed to parse series time: %s", value)
}

// workersLogsFlattenEvent flattens the raw JSON of a Workers Observability
// event into a map with dot separated keys (e.g.
// "$workers.event.request.cf.country"), so that all nested fields of an event
// are shown as separate labels in Grafana.
func workersLogsFlattenEvent(raw string) (json.RawMessage, error) {
	var event map[string]any
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return nil, err
	}

	flattened := make(map[string]string)
	workersLogsFlattenValue("", event, flattened)

	return json.Marshal(flattened)
}

// workersLogsFlattenValue adds the given value to the flattened map. Objects
// and arrays are flattened recursively, where the keys / indices are appended
// to the key with a dot as separator. Null values are skipped.
func workersLogsFlattenValue(key string, value any, flattened map[string]string) {
	switch v := value.(type) {
	case map[string]any:
		for k, val := range v {
			if key == "" {
				workersLogsFlattenValue(k, val, flattened)
			} else {
				workersLogsFlattenValue(key+"."+k, val, flattened)
			}
		}
	case []any:
		for i, val := range v {
			workersLogsFlattenValue(key+"."+strconv.Itoa(i), val, flattened)
		}
	case string:
		flattened[key] = v
	case float64:
		flattened[key] = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		flattened[key] = strconv.FormatBool(v)
	case nil:
	default:
		flattened[key] = fmt.Sprintf("%v", v)
	}
}

func (c *client) GetWorkersLogs(ctx context.Context, accountId string, filters []models.QueryModelMetricsFilter, timeFrom, timeTo time.Time, limit int64) backend.DataResponse {
	res, err := c.client.Workers.Observability.Telemetry.Query(ctx, workers.ObservabilityTelemetryQueryParams{
		AccountID: cloudflare.F(accountId),
		QueryID:   cloudflare.F(workersLogsQueryID),
		Dry:       cloudflare.F(true),
		Timeframe: cloudflare.F(workers.ObservabilityTelemetryQueryParamsTimeframe{
			From: cloudflare.F(float64(timeFrom.UnixMilli())),
			To:   cloudflare.F(float64(timeTo.UnixMilli())),
		}),
		View:  cloudflare.F(workers.ObservabilityTelemetryQueryParamsViewEvents),
		Limit: cloudflare.F(float64(limit)),
		Parameters: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParameters{
			Datasets:          cloudflare.F(workersLogsDatasets),
			FilterCombination: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFilterCombinationAnd),
			Filters:           cloudflare.F(workersLogsFiltersToTelemetry(filters)),
			Limit:             cloudflare.F(limit),
		}),
	})
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	var timestamps []time.Time
	var bodys []string
	var severities []string
	var labels []json.RawMessage

	for _, e := range res.Events.Events {
		timestamp := time.UnixMilli(e.Timestamp)

		body := e.Metadata.Message
		if body == "" {
			body = e.Metadata.Error
		}
		if body == "" {
			body = fmt.Sprintf("scriptName=%s eventType=%s outcome=%s", e.Workers.ScriptName, e.Workers.EventType, e.Workers.Outcome)
		}

		label, err := workersLogsFlattenEvent(e.JSON.RawJSON())
		if err != nil {
			c.logger.Error("Failed to flatten event", "error", err)
			label = json.RawMessage(e.JSON.RawJSON())
		}

		timestamps = append(timestamps, timestamp)
		bodys = append(bodys, body)
		severities = append(severities, workersLogsLevelToSeverity(e.Metadata.Level))
		labels = append(labels, label)
	}

	frame := data.NewFrame(
		"workersLogs",
		data.NewField("timestamp", nil, timestamps),
		data.NewField("body", nil, bodys),
		data.NewField("severity", nil, severities),
		data.NewField("labels", nil, labels),
	)

	frame.SetMeta(&data.FrameMeta{
		PreferredVisualization: data.VisTypeLogs,
		Type:                   data.FrameTypeLogLines,
	})

	var response backend.DataResponse
	response.Frames = append(response.Frames, frame)

	return response
}

func (c *client) GetWorkersLogsVolumes(ctx context.Context, accountId string, filters []models.QueryModelMetricsFilter, timeFrom, timeTo time.Time) backend.DataResponse {
	res, err := c.client.Workers.Observability.Telemetry.Query(ctx, workers.ObservabilityTelemetryQueryParams{
		AccountID: cloudflare.F(accountId),
		QueryID:   cloudflare.F(workersLogsQueryID),
		Dry:       cloudflare.F(true),
		Timeframe: cloudflare.F(workers.ObservabilityTelemetryQueryParamsTimeframe{
			From: cloudflare.F(float64(timeFrom.UnixMilli())),
			To:   cloudflare.F(float64(timeTo.UnixMilli())),
		}),
		// The chart option must be set to true, since the time series data
		// for the calculations view is only included in the response if it
		// is set.
		Chart: cloudflare.F(true),
		View:  cloudflare.F(workers.ObservabilityTelemetryQueryParamsViewCalculations),
		Parameters: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParameters{
			Calculations: cloudflare.F([]workers.ObservabilityTelemetryQueryParamsParametersCalculation{{
				Operator: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersCalculationsOperatorCount),
			}}),
			Datasets:          cloudflare.F(workersLogsDatasets),
			FilterCombination: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFilterCombinationAnd),
			Filters:           cloudflare.F(workersLogsFiltersToTelemetry(filters)),
			GroupBys: cloudflare.F([]workers.ObservabilityTelemetryQueryParamsParametersGroupBy{{
				Type:  cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersGroupBysTypeString),
				Value: cloudflare.F("$metadata.level"),
			}}),
		}),
	})
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	// Collect the count of log lines for each severity and timestamp. Since
	// multiple levels can be mapped to the same severity (e.g. "log" and
	// "info"), the counts are summed up.
	severityValues := make(map[string]map[time.Time]float64)

	for _, calculation := range res.Calculations {
		for _, series := range calculation.Series {
			timestamp, err := workersLogsParseSeriesTime(series.Time)
			if err != nil {
				c.logger.Error("Failed to parse timestamp", "error", err)
				continue
			}

			for _, d := range series.Data {
				severity := "info"
				for _, g := range d.Groups {
					if g.Key == "$metadata.level" {
						severity = workersLogsLevelToSeverity(workersLogsUnionToString(g.Value))
					}
				}

				if _, ok := severityValues[severity]; !ok {
					severityValues[severity] = make(map[time.Time]float64)
				}
				severityValues[severity][timestamp] += d.Value
			}
		}
	}

	var response backend.DataResponse

	for severity, values := range severityValues {
		var timestamps []time.Time
		for timestamp := range values {
			timestamps = append(timestamps, timestamp)
		}
		sort.Slice(timestamps, func(i, j int) bool {
			return timestamps[i].Before(timestamps[j])
		})

		var sortedValues []float64
		for _, timestamp := range timestamps {
			sortedValues = append(sortedValues, values[timestamp])
		}

		frame := data.NewFrame(
			severity,
			data.NewField("Time", nil, timestamps),
			data.NewField(severity, map[string]string{"level": severity}, sortedValues),
		)

		response.Frames = append(response.Frames, frame)
	}

	return response
}

// GetWorkersLogsValues returns the distinct values of the given field. Since
// the Workers Observability values endpoint only supports a limited set of
// fields, the values are instead determined via a calculations query which is
// grouped by the selected field (similar to GetWorkersLogsVolumes). This way
// any key of an event can be used as field (e.g.
// "$workers.event.request.cf.country").
func (c *client) GetWorkersLogsValues(ctx context.Context, accountId, field string, timeFrom, timeTo time.Time, limit int64) backend.DataResponse {
	res, err := c.client.Workers.Observability.Telemetry.Query(ctx, workers.ObservabilityTelemetryQueryParams{
		AccountID: cloudflare.F(accountId),
		QueryID:   cloudflare.F(workersLogsQueryID),
		Dry:       cloudflare.F(true),
		Timeframe: cloudflare.F(workers.ObservabilityTelemetryQueryParamsTimeframe{
			From: cloudflare.F(float64(timeFrom.UnixMilli())),
			To:   cloudflare.F(float64(timeTo.UnixMilli())),
		}),
		// The chart option must be set to true, since the time series data
		// for the calculations view is only included in the response if it
		// is set.
		Chart: cloudflare.F(true),
		View:  cloudflare.F(workers.ObservabilityTelemetryQueryParamsViewCalculations),
		Parameters: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParameters{
			Calculations: cloudflare.F([]workers.ObservabilityTelemetryQueryParamsParametersCalculation{{
				Operator: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersCalculationsOperatorCount),
			}}),
			Datasets:          cloudflare.F(workersLogsDatasets),
			FilterCombination: cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersFilterCombinationAnd),
			GroupBys: cloudflare.F([]workers.ObservabilityTelemetryQueryParamsParametersGroupBy{{
				Type:  cloudflare.F(workers.ObservabilityTelemetryQueryParamsParametersGroupBysTypeString),
				Value: cloudflare.F(field),
			}}),
			Limit: cloudflare.F(limit),
		}),
	})
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	var response backend.DataResponse

	// Collect the distinct values of the field across all series. The value of
	// each group is used as the name of a frame, so that it is returned as a
	// filter value in the frontend.
	seen := make(map[string]struct{})

	for _, calculation := range res.Calculations {
		for _, series := range calculation.Series {
			for _, d := range series.Data {
				for _, g := range d.Groups {
					if g.Key != field {
						continue
					}

					value := workersLogsUnionToString(g.Value)
					if _, ok := seen[value]; ok {
						continue
					}
					seen[value] = struct{}{}

					response.Frames = append(response.Frames, data.NewFrame(value))

					if limit > 0 && int64(len(seen)) >= limit {
						return response
					}
				}
			}
		}
	}

	return response
}
