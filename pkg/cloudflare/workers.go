package cloudflare

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type WorkersInvocationsAggregateResponse struct {
	Viewer struct {
		Accounts []struct {
			WorkersInvocationsAdaptive []struct {
				Dimensions map[string]any     `json:"dimensions"`
				Sum        map[string]float64 `json:"sum"`
				Quantiles  map[string]float64 `json:"quantiles"`
			} `json:"workersInvocationsAdaptive"`
		} `json:"accounts"`
	} `json:"viewer"`
}

func (c *client) GetWorkersInvocationsAggregate(ctx context.Context, accountId, metricName, aggregation, filters, dimensions, orderBy, legend string, limit int64, timeTo time.Time) backend.DataResponse {
	metricName = strings.TrimPrefix(metricName, "workersInvocations_")

	var aggregationGraphQL string
	var valueKey string
	switch {
	case aggregation == "sum":
		valueKey = metricName
		aggregationGraphQL = fmt.Sprintf("sum { %s }", valueKey)
	case strings.HasPrefix(aggregation, "P"):
		valueKey = metricName + aggregation
		aggregationGraphQL = fmt.Sprintf("quantiles { %s }", valueKey)
	default:
		c.logger.Error("Unsupported aggregation", "aggregation", aggregation)
		return backend.ErrorResponseWithErrorSource(fmt.Errorf("unsupported aggregation: %s", aggregation))
	}

	query := fmt.Sprintf(`{
		viewer {
			accounts(filter: {accountTag: "%s"}) {
				workersInvocationsAdaptive(
					%s
					limit: %d
					orderBy: [%s]
				) {
					%s
					%s
				}
			}
		}
	}`, accountId, filters, limit, orderBy, aggregationGraphQL, dimensions)

	res, err := graphQLRequest[WorkersInvocationsAggregateResponse](ctx, c.client, query)
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	frameData := make(map[string]FrameData)

	for _, a := range res.Viewer.Accounts {
		for i, r := range a.WorkersInvocationsAdaptive {
			var keys []string
			timestamp := timeTo
			var labels = make(map[string]string)

			for k, v := range r.Dimensions {
				if strings.HasPrefix(k, "datetime") {
					t, err := time.Parse(time.RFC3339, v.(string))
					if err != nil {
						c.logger.Error("Failed to parse timestamp", "error", err)
						continue
					}
					timestamp = t
				} else if k == "date" {
					t, err := time.Parse(time.DateOnly, v.(string))
					if err != nil {
						c.logger.Error("Failed to parse timestamp", "error", err)
						continue
					}
					timestamp = t
				} else {
					keys = append(keys, fmt.Sprintf("%s=\"%v\"", k, v))
					labels[k] = fmt.Sprintf("%v", v)
				}
			}
			slices.Sort(keys)
			key := fmt.Sprintf("%s{%s}", valueKey, strings.Join(keys, ","))

			var value float64
			if aggregation == "sum" {
				value = r.Sum[valueKey]
			} else {
				value = r.Quantiles[valueKey]
			}

			if f, ok := frameData[key]; ok {
				f.Timestamps = append(f.Timestamps, timestamp)
				f.Values = append(f.Values, value)
				frameData[key] = f
			} else {
				frameData[key] = FrameData{
					Index:      i,
					Name:       key,
					Timestamps: []time.Time{timestamp},
					Values:     []float64{value},
					Labels:     labels,
				}
			}
		}
	}

	var response backend.DataResponse

	// Convert the FrameData map to a slice and sort it by the original index to
	// maintain the order of the groups as returned by the API.
	var frameDataSlice []FrameData
	for _, v := range frameData {
		frameDataSlice = append(frameDataSlice, v)
	}
	sort.Slice(frameDataSlice, func(i, j int) bool {
		return frameDataSlice[i].Index < frameDataSlice[j].Index
	})

	for _, v := range frameDataSlice {
		name := parseLegend(v.Name, legend, v.Labels)

		frame := data.NewFrame(
			name,
			data.NewField("Time", nil, v.Timestamps),
			data.NewField(valueKey, v.Labels, v.Values),
		)

		frame.Fields[1].Config = &data.FieldConfig{DisplayNameFromDS: name}

		frame.SetMeta(&data.FrameMeta{
			PreferredVisualization: data.VisTypeGraph,
			Type:                   data.FrameTypeTimeSeriesMulti,
			TypeVersion:            data.FrameTypeVersion{0, 1},
		})

		response.Frames = append(response.Frames, frame)
	}

	return response
}
