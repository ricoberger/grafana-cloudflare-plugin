package cloudflare

import (
	"context"
	"fmt"
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

func FiltersToGraphQL(filters []models.QueryModelMetricsFilter) string {
	var filterStrings []string

	for _, f := range filters {
		if f.Field != "-" {
			v, err := strconv.ParseFloat(f.Value, 64)
			if err == nil {
				f.Value = fmt.Sprintf("%f", v)
			} else {
				f.Value = fmt.Sprintf(`"%s"`, f.Value)
			}

			filterStrings = append(filterStrings, fmt.Sprintf("%s%s: %s", f.Field, filterOperators[f.Operator], f.Value))
		}
	}
	return fmt.Sprintf("filter: { %s }", strings.Join(filterStrings, ", "))
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
