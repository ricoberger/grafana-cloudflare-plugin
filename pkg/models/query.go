package models

type QueryType string

const (
	QueryTypeZones   = "zones"
	QueryTypeMetrics = "metrics"
)

type QueryModelMetrics struct {
	Name       string                    `json:"name"`
	Zone       string                    `json:"zone"`
	Filters    []QueryModelMetricsFilter `json:"filters"`
	Dimensions []string                  `json:"dimensions"`
	Legend     string                    `json:"legend"`
	Limit      int64                     `json:"limit"`
}

type QueryModelMetricsFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
