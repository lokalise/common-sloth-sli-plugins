package availability

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const (
	SLIPluginVersion = "prometheus/v1"
	SLIPluginID      = "lokalise/nginx-http/latency"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
1 - ((
	sum(
		rate(nginx_ingress_controller_request_duration_seconds_bucket{ {{ .filter }}exported_service=~"{{ .serviceName }}", le="{{ .bucket }}" }[{{"{{ .window }}"}}])
	)
	/
	(sum(
		rate(nginx_ingress_controller_request_duration_seconds_count{ {{ .filter }}exported_service=~"{{ .serviceName }}" }[{{"{{ .window }}"}}])
	) > 0)
) OR on() vector(1))
`))

// SLIPlugin will return a query that will return the availability error based on traefik V1 service metrics.
func SLIPlugin(ctx context.Context, meta, labels, options map[string]string) (string, error) {
	service, err := getServiceName(options)
	if err != nil {
		return "", fmt.Errorf("could not get service name: %w", err)
	}

	bucket, err := getBucket(options)
	if err != nil {
		return "", fmt.Errorf(`could not get bucket: %w`, err)
	}

	var b bytes.Buffer
	data := map[string]string{
		"filter":      getFilter(options),
		"bucket":      bucket,
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

func getBucket(options map[string]string) (string, error) {
	bucket := options["bucket"]
	if bucket == "" {
		return "", fmt.Errorf(`"bucket" option is required`)
	}

	_, err := strconv.ParseFloat(bucket, 64)
	if err != nil {
		return "", fmt.Errorf("not a valid bucket, can't parse to float64: %w", err)
	}

	return bucket, nil
}
