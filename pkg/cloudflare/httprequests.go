package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type HttpRequestsResponse struct {
	Viewer struct {
		Zones []struct {
			HttpRequestsAdaptive []map[string]any `json:"httpRequestsAdaptive"`
		} `json:"zones"`
	} `json:"viewer"`
}

func (c *client) GetHTTPRequests(ctx context.Context, zoneId, filters string, limit int64) backend.DataResponse {
	query := fmt.Sprintf(`{
		viewer {
			zones(filter: {zoneTag: "%s"}) {
				httpRequestsAdaptive(
					%s
					limit: %d
					orderBy: [datetime_DESC]
				) {
					datetime
					cacheStatus
					clientASNDescription
					clientAsn
					clientCountryName
					clientDeviceType
					clientIP
					clientRefererHost
					clientRequestHTTPHost
					clientRequestHTTPMethodName
					clientRequestHTTPProtocol
					clientRequestPath
					clientRequestQuery
					clientRequestReferer
					clientRequestScheme
					clientSSLProtocol
					coloCode
					edgeDnsResponseTimeMs
					edgeResponseContentTypeName
					edgeResponseStatus
					edgeTimeToFirstByteMs
					originASN
					originASNDescription
					originIP
					originResponseDurationMs
					originResponseStatus
					rayName
					requestSource
					upperTierColoName
					userAgent
					userAgentBrowser
					userAgentOS
					verifiedBotCategory
					wafAttackScore
					wafAttackScoreClass
					wafRceAttackScore
					wafSqliAttackScore
					wafXssAttackScore
				}
			}
		}
	}`, zoneId, filters, limit)

	res, err := graphQLRequest[HttpRequestsResponse](ctx, c.client, query)
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	var timestamps []time.Time
	var bodys []string
	var severities []string
	var labels []json.RawMessage

	for _, z := range res.Viewer.Zones {
		for _, r := range z.HttpRequestsAdaptive {
			timestamp, err := time.Parse(time.RFC3339, r["datetime"].(string))
			if err != nil {
				c.logger.Error("Failed to parse timestamp", "error", err)
				continue
			}

			status, ok := r["edgeResponseStatus"].(float64)
			if !ok {
				c.logger.Error("Failed to parse status code")
				continue
			}

			method, ok := r["clientRequestHTTPMethodName"].(string)
			if !ok {
				c.logger.Error("Failed to parse HTTP method")
				continue
			}

			scheme, ok := r["clientRequestScheme"].(string)
			if !ok {
				c.logger.Error("Failed to parse request scheme")
				continue
			}

			host, ok := r["clientRequestHTTPHost"].(string)
			if !ok {
				c.logger.Error("Failed to parse request host")
				continue
			}

			path, ok := r["clientRequestPath"].(string)
			if !ok {
				c.logger.Error("Failed to parse request path")
				continue
			}

			severity := ""
			switch {
			case status < 300:
				severity = "info"
			case status >= 300 && status < 400:
				severity = "warning"
			case status >= 400 && status < 500:
				severity = "error"
			case status >= 500:
				severity = "critical"
			}

			jsonData, err := json.Marshal(r)
			if err != nil {
				c.logger.Error("Failed to marshal log labels", "error", err)
				continue
			}
			var label json.RawMessage = jsonData

			timestamps = append(timestamps, timestamp)
			bodys = append(bodys, fmt.Sprintf("status=%.0f method=%s url=%s://%s%s", status, method, scheme, host, path))
			severities = append(severities, severity)
			labels = append(labels, label)
		}
	}

	frame := data.NewFrame(
		"httpRequests",
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

type HttpRequestsAggregateResponse struct {
	Viewer struct {
		Zones []struct {
			HttpRequestsAdaptiveGroups []struct {
				Dimensions map[string]any     `json:"dimensions"`
				Sum        map[string]float64 `json:"sum"`
			} `json:"httpRequestsAdaptiveGroups"`
			HttpRequestsOverviewAdaptiveGroups []struct {
				Dimensions map[string]any     `json:"dimensions"`
				Sum        map[string]float64 `json:"sum"`
			} `json:"httpRequestsOverviewAdaptiveGroups"`
		} `json:"zones"`
	} `json:"viewer"`
}

func (c *client) GetHTTPRequestsAggregate(ctx context.Context, zoneId, metricName, filters, dimensions, legend string, limit int64, timeTo time.Time) backend.DataResponse {
	var group string
	if strings.HasPrefix(metricName, "httpRequests_overview_") {
		group = "httpRequestsOverviewAdaptiveGroups"
		metricName = strings.TrimPrefix(metricName, "httpRequests_overview_")
	} else {
		group = "httpRequestsAdaptiveGroups"
		metricName = strings.TrimPrefix(metricName, "httpRequests_")
	}

	query := fmt.Sprintf(`{
		viewer {
			zones(filter: {zoneTag: "%s"}) {
				%s(
					%s
					limit: %d
				) {
					sum { %s }
					%s
				}
			}
		}
	}`, zoneId, group, filters, limit, metricName, dimensions)

	res, err := graphQLRequest[HttpRequestsAggregateResponse](ctx, c.client, query)
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	frameData := make(map[string]FrameData)

	for _, z := range res.Viewer.Zones {
		groups := z.HttpRequestsAdaptiveGroups
		if len(groups) == 0 {
			groups = z.HttpRequestsOverviewAdaptiveGroups
		}

		for _, r := range groups {
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
			key := fmt.Sprintf("%s{%s}", metricName, strings.Join(keys, ","))

			if f, ok := frameData[key]; ok {
				f.Timestamps = append(f.Timestamps, timestamp)
				f.Values = append(f.Values, r.Sum[metricName])
				frameData[key] = f
			} else {
				frameData[key] = FrameData{
					Name:       key,
					Timestamps: []time.Time{timestamp},
					Values:     []float64{r.Sum[metricName]},
					Labels:     labels,
				}
			}
		}
	}

	var response backend.DataResponse

	for _, v := range frameData {
		name := parseLegend(v.Name, legend, v.Labels)

		frame := data.NewFrame(
			name,
			data.NewField("Time", nil, v.Timestamps),
			data.NewField(metricName, v.Labels, v.Values),
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
