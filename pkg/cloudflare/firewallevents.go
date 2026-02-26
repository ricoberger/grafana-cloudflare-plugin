package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type FirewallEventsResponse struct {
	Viewer struct {
		Zones []struct {
			FirewallEventsAdaptive []map[string]any `json:"firewallEventsAdaptive"`
		} `json:"zones"`
	} `json:"viewer"`
}

func (c *client) GetFirewallEvents(ctx context.Context, zoneId, filters string, limit int64) backend.DataResponse {
	query := fmt.Sprintf(`{
		viewer {
			zones(filter: {zoneTag: "%s"}) {
				firewallEventsAdaptive(
					%s
					limit: %d
					orderBy: [datetime_DESC]
				) {
					datetime
					action
					clientASNDescription
					clientAsn
					clientCountryName
					clientIP
					clientIPClass
					clientRefererHost
					clientRefererPath
					clientRefererQuery
					clientRefererScheme
					clientRequestHTTPHost
					clientRequestHTTPMethodName
					clientRequestHTTPProtocol
					clientRequestPath
					clientRequestQuery
					clientRequestScheme
					description
					edgeColoName
					edgeResponseStatus
					kind
					matchIndex
					originResponseStatus
					originatorRayName
					rayName
					ref
					ruleId
					rulesetId
					source
					userAgent
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

	res, err := graphQLRequest[FirewallEventsResponse](ctx, c.client, query)
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	var timestamps []time.Time
	var bodys []string
	var severities []string
	var labels []json.RawMessage

	for _, z := range res.Viewer.Zones {
		for _, r := range z.FirewallEventsAdaptive {
			timestamp, err := time.Parse(time.RFC3339, r["datetime"].(string))
			if err != nil {
				c.logger.Error("Failed to parse timestamp", "error", err)
				continue
			}

			sourceIP, ok := r["clientIP"].(string)
			if !ok {
				c.logger.Error("Failed to parse source ip")
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

			rayID, ok := r["rayName"].(string)
			if !ok {
				c.logger.Error("Failed to parse ray id")
				continue
			}

			analyses, ok := r["wafAttackScoreClass"].(string)
			if !ok {
				c.logger.Error("Failed to parse analyses")
				continue
			}

			severity := ""
			switch analyses {
			case "clean":
				severity = "info"
			case "likely_clean":
				severity = "warning"
			case "likely_attack":
				severity = "error"
			case "attack":
				severity = "critical"
			}

			jsonData, err := json.Marshal(r)
			if err != nil {
				c.logger.Error("Failed to marshal log labels", "error", err)
				continue
			}
			var label json.RawMessage = jsonData

			timestamps = append(timestamps, timestamp)
			bodys = append(bodys, fmt.Sprintf("sourceIP=%s host=%s path=%s rayID=%s analyses=%s", sourceIP, host, path, rayID, analyses))
			severities = append(severities, severity)
			labels = append(labels, label)
		}
	}

	frame := data.NewFrame(
		"firewallEvents",
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

func (c *client) GetFirewallEventsVolumes(ctx context.Context, zoneId, filtersInfo, filtersWarning, filtersError, filtersCritical string) backend.DataResponse {
	volumes := []Volume{{
		Name:      "info",
		Filter:    filtersInfo,
		Dimension: "datetime",
	}, {
		Name:      "warning",
		Filter:    filtersWarning,
		Dimension: "datetime",
	}, {
		Name:      "error",
		Filter:    filtersError,
		Dimension: "datetime",
	}, {
		Name:      "critical",
		Filter:    filtersCritical,
		Dimension: "datetime",
	}}

	var response backend.DataResponse
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(volumes))

	for _, volume := range volumes {
		go func(volume Volume) {
			defer wg.Done()

			query := fmt.Sprintf(`{
				viewer {
					zones(filter: {zoneTag: "%s"}) {
						firewallEventsAdaptiveGroups(
							%s
							limit: 100
						) {
							count
							dimensions { %s }
						}
					}
				}
			}`, zoneId, volume.Filter, volume.Dimension)

			res, err := graphQLRequest[FirewallEventsAggregateResponse](ctx, c.client, query)
			if err != nil {
				c.logger.Error("Request failed", "error", err)
				return
			}

			var timestamps []time.Time
			var values []float64

			for _, z := range res.Viewer.Zones {
				for _, r := range z.FirewallEventsAdaptiveGroups {
					t, err := time.Parse(time.RFC3339, r.Dimensions[volume.Dimension].(string))
					if err != nil {
						c.logger.Error("Failed to parse timestamp", "error", err)
						continue
					}
					timestamps = append(timestamps, t)
					values = append(values, r.Count)
				}
			}

			frame := data.NewFrame(
				volume.Name,
				data.NewField("Time", nil, timestamps),
				data.NewField(volume.Name, map[string]string{"level": volume.Name}, values),
			)

			mu.Lock()
			response.Frames = append(response.Frames, frame)
			mu.Unlock()
		}(volume)
	}

	wg.Wait()
	return response
}

type FirewallEventsAggregateResponse struct {
	Viewer struct {
		Zones []struct {
			FirewallEventsAdaptiveGroups []struct {
				Dimensions map[string]any `json:"dimensions"`
				Count      float64        `json:"count"`
			} `json:"firewallEventsAdaptiveGroups"`
		} `json:"zones"`
	} `json:"viewer"`
}

func (c *client) GetFirewallEventsAggregate(ctx context.Context, zoneId, filters, dimensions, orderBy, legend string, limit int64, timeTo time.Time) backend.DataResponse {
	query := fmt.Sprintf(`{
		viewer {
			zones(filter: {zoneTag: "%s"}) {
				firewallEventsAdaptiveGroups(
					%s
					limit: %d
					orderBy: [%s]
				) {
					count
					%s
				}
			}
		}
	}`, zoneId, filters, limit, orderBy, dimensions)

	res, err := graphQLRequest[FirewallEventsAggregateResponse](ctx, c.client, query)
	if err != nil {
		return backend.ErrorResponseWithErrorSource(err)
	}

	frameData := make(map[string]FrameData)

	for _, z := range res.Viewer.Zones {
		for i, r := range z.FirewallEventsAdaptiveGroups {
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
			key := fmt.Sprintf("firewallEvents_events{%s}", strings.Join(keys, ","))

			if f, ok := frameData[key]; ok {
				f.Timestamps = append(f.Timestamps, timestamp)
				f.Values = append(f.Values, r.Count)
				frameData[key] = f
			} else {
				frameData[key] = FrameData{
					Index:      i,
					Name:       key,
					Timestamps: []time.Time{timestamp},
					Values:     []float64{r.Count},
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
			data.NewField("firewallEvents_events", v.Labels, v.Values),
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
