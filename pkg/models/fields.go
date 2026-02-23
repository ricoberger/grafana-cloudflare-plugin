package models

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Fields []*data.Field

func (f *Fields) Add(name string, labels data.Labels, values any, config ...*data.FieldConfig) *data.Field {
	field := data.NewField(name, labels, values)
	if len(config) > 0 {
		field.SetConfig(config[0])
	}
	*f = append(*f, field)
	return field
}
