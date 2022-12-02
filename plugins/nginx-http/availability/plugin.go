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
	SLIPluginID      = "lokalise/nginx-http/availability"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
(
	sum(
		rate(nginx_ingress_controller_request_duration_seconds_count{ {{ .filter }}exported_service=~"{{ .serviceName }}", status=~"(5..|429|431)" }[{{"{{ .window }}"}}])
	)
	/
	(sum(
		rate(nginx_ingress_controller_request_duration_seconds_count{ {{ .filter }}exported_service=~"{{ .serviceName }}" }[{{"{{ .window }}"}}])
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
		"filter":      getFilter(options),
		"serviceName": service,
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
