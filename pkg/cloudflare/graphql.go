package cloudflare

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ricoberger/grafana-cloudflare-plugin/pkg/models"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
)

type GraphQLResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func graphQLRequest[T any](ctx context.Context, client *cloudflare.Client, query string) (T, error) {
	var result GraphQLResponse[T]

	body := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	err := client.Post(ctx, "/graphql", body, &result)
	if err != nil {
		return result.Data, err
	}

	if len(result.Errors) > 0 {
		var messages []string
		for _, e := range result.Errors {
			messages = append(messages, e.Message)
		}
		return result.Data, fmt.Errorf("GraphQL errors: %s", strings.Join(messages, "; "))
	}

	return result.Data, nil
}

var filterOperators = map[string]string{
	"=":  "",
	"!=": "_neq",
	">":  "_gt",
	"<":  "_lt",
	">=": "_geq",
	"<=": "_leq",
}

func FiltersToGraphQL(timeFrom, timeTo, filter string, filters []models.QueryModelMetricsFilter, additionalFilter string) string {
	// Filter out all filters where the field value is "-", as this is used to
	// indicate that the filter should be ignored. This allows us to display a
	// new empty filter in the UI without it affecting the query.
	filters = slices.Collect(func(yield func(models.QueryModelMetricsFilter) bool) {
		for _, f := range filters {
			if f.Field != "-" {
				if !yield(f) {
					return
				}
			}
		}
	})

	// If there is no filter and no filters, return a filter with only the time
	// range.
	if filter == "" && len(filters) == 0 {
		if additionalFilter != "" {
			return fmt.Sprintf("filter: { datetime_geq: \"%s\", datetime_leq: \"%s\", %s }", timeFrom, timeTo, additionalFilter)
		}
		return fmt.Sprintf("filter: { datetime_geq: \"%s\", datetime_leq: \"%s\" }", timeFrom, timeTo)
	}

	if filter != "" {
		if additionalFilter != "" {
			return fmt.Sprintf("filter: { AND: [{ datetime_geq: \"%s\", datetime_leq: \"%s\" }, { %s }  %s ] }", timeFrom, timeTo, additionalFilter, strings.ReplaceAll(strings.ReplaceAll(filter, "\n", ""), "\r", ""))
		}
		return fmt.Sprintf("filter: { AND: [{ datetime_geq: \"%s\", datetime_leq: \"%s\" },  %s ] }", timeFrom, timeTo, strings.ReplaceAll(strings.ReplaceAll(filter, "\n", ""), "\r", ""))
	}

	// If there is no filter, but filters we build the GraphQL filter string
	// based on the filters. We also need to convert the filter operators to the
	// GraphQL format, which is done using the filterOperators map. We also need
	// to convert the filter values to the correct format, which is done by
	// trying to parse the value as a float. If it can be parsed as a float, we
	// format it as a float. If it cannot be parsed as a float, we format it as
	// a string.
	if len(filters) > 0 {
		var filterStrings []string
		for _, f := range filters {
			v, err := strconv.ParseFloat(f.Value, 64)
			if err == nil {
				f.Value = fmt.Sprintf("%f", v)
			} else {
				f.Value = fmt.Sprintf(`"%s"`, f.Value)
			}
			filterStrings = append(filterStrings, fmt.Sprintf("%s%s: %s", f.Field, filterOperators[f.Operator], f.Value))
		}
		if additionalFilter != "" {
			return fmt.Sprintf("filter: { AND: [{ datetime_geq: \"%s\", datetime_leq: \"%s\" }, { %s }, { %s }] }", timeFrom, timeTo, additionalFilter, strings.Join(filterStrings, ", "))
		}
		return fmt.Sprintf("filter: { AND: [{ datetime_geq: \"%s\", datetime_leq: \"%s\" }, { %s }] }", timeFrom, timeTo, strings.Join(filterStrings, ", "))
	}

	return ""
}

func DimensionsToGraphQL(dimensions []string) string {
	if len(dimensions) == 0 {
		return ""
	}

	return fmt.Sprintf("dimensions { %s }", strings.Join(dimensions, ", "))
}

func OrderByToGraphQL(orderBy []string) string {
	if len(orderBy) == 0 {
		return ""
	}

	return strings.Join(orderBy, ", ")
}
