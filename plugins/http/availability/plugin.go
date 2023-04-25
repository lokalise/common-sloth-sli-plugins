package availability

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

const (
	SLIPluginVersion = "prometheus/v1"
	SLIPluginID      = "lokalise/http/availability"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
(
	sum(
		rate({{ .metric_name }}_count{ {{ .filter }}service=~"{{ .serviceName }}", route=~"{{ .route }}", status=~"{{ .status }}" }[{{"{{ .window }}"}}])
	)
	/
	(sum(
		rate({{ .metric_name }}_count{ {{ .filter }}service=~"{{ .serviceName }}", route=~"{{ .route }}"}[{{"{{ .window }}"}}])
	) > 0)
) OR on() vector(0)
`))

// SLIPlugin will return a query that will return the availability error based on traefik V1 service metrics.
func SLIPlugin(ctx context.Context, meta, labels, options map[string]string) (string, error) {
	service, err := getServiceName(options)
	if err != nil {
		return "", fmt.Errorf("could not get service name: %w", err)
	}

	var b bytes.Buffer
	data := map[string]string{
		"metric_name": getMetricName(options),
		"filter":      getFilter(options),
		"serviceName": service,
		"status":      getStatus(options),
		"route":       getRoute(options),
	}
	err = queryTpl.Execute(&b, data)
	if err != nil {
		return "", fmt.Errorf("could not render query template: %w", err)
	}

	return b.String(), nil
}

func getFilter(options map[string]string) string {
	filter := options["filter"]
	filter = strings.Trim(filter, "{},")
	if filter != "" {
		filter += ","
	}

	return filter
}

func getServiceName(options map[string]string) (string, error) {
	service := options["service_name_regex"]
	service = strings.TrimSpace(service)

	if service == "" {
		return "", fmt.Errorf("service name is required")
	}

	_, err := regexp.Compile(service)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %w", err)
	}

	return service, nil
}

func getRoute(options map[string]string) string {
	route := options["route_regex"]
	route = strings.TrimSpace(route)

	if route == "" {
		route = ".*"
	}

	return route
}

func getStatus(options map[string]string) string {
	status := options["status_regex"]
	status = strings.TrimSpace(status)

	if status == "" {
		status = "(5..|429|431)"
	}

	return status
}

func getMetricName(options map[string]string) string {
	metricName := options["metric_name"]
	if metricName == "" {
		metricName = "http_request_duration_seconds"
	}

	return metricName
}
