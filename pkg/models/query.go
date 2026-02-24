package models

type QueryType string

const (
	QueryTypeZones   = "zones"
	QueryTypeMetrics = "metrics"
)

type QueryModelMetrics struct {
	Name        string                    `json:"name"`
	Aggregation string                    `json:"aggregation"`
	Zone        string                    `json:"zone"`
	Filter      string                    `json:"filter"`
	Filters     []QueryModelMetricsFilter `json:"filters"`
	Dimensions  []string                  `json:"dimensions"`
	OrderBy     []string                  `json:"orderBy"`
	Legend      string                    `json:"legend"`
	Limit       int64                     `json:"limit"`
}

type QueryModelMetricsFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
