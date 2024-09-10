package uptime

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
	SLIPluginID      = "lokalise/uptime"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
sum(count_over_time(({{ .metricName }}{ {{ .additionalLabels }}{{ .ingressLabelName }}=~"{{ .ingressLabelValue }}" } == 0)[{{"{{ .window }}"}}:])) or vector(0)
/
sum(count_over_time(({{ .metricName }}{ {{ .additionalLabels }}{{ .ingressLabelName }}=~"{{ .ingressLabelValue }}" })[{{"{{ .window }}"}}:]))
`))

// SLIPlugin will return a query that will return the availability error based on traefik V1 ingress metrics.
func SLIPlugin(ctx context.Context, meta, labels, options map[string]string) (string, error) {
	metricName, err := getMetricName(options)
	ingressLabelName, err := getIngressLabelName(options)
	ingressLabelValue, err := getIngressLabelValue(options)

	if err != nil {
		return "", fmt.Errorf("Error parsing options: %w", err)
	}

	var b bytes.Buffer
	data := map[string]string{
		"metricName":               metricName,
		"ingressLabelName":         ingressLabelName,
		"ingressLabelValue":        ingressLabelValue,
		"additionalLabels":         getAdditionalLabels(options),
	}
	err = queryTpl.Execute(&b, data)
	if err != nil {
		return "", fmt.Errorf("could not render query template: %w", err)
	}

	return b.String(), nil
}

func getAdditionalLabels(options map[string]string) string {
	labels := options["additionalLabels"]
	labels = strings.Trim(labels, "{},")

	if labels != "" {
		labels += ", "
	}

	return labels
}

func getIngressLabelName(options map[string]string) (string, error) {
	label := options["ingressLabelName"]
	label = strings.TrimSpace(label)

	if label == "" {
		return "", fmt.Errorf("'ingressLabelName' name is required")
	}

	return label, nil
}

func getIngressLabelValue(options map[string]string) (string, error) {
	value := options["ingressLabelValue"]
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf("'ingressLabelValue' is required")
	}

	_, err := regexp.Compile(value)
	if err != nil {
		return "", fmt.Errorf("invalid regex for 'ingressLabelValue': %w", err)
	}

	return value, nil
}

func getMetricName(options map[string]string) (string, error) {
	metricName := options["metricName"]
	if metricName == "" {
		return "", fmt.Errorf("'metricName' is required")
	}

	return metricName, nil
}
