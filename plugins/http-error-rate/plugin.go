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
	SLIPluginID      = "lokalise/http-error-rate"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
(
	(
		sum(
			rate({{ .metricName }}{ {{ .additionalLabels }}{{ .serviceLabelName }}=~"{{ .serviceLabelValue }}", {{ .errorLabelName }}=~"{{ .errorLabelValue }}"}[{{"{{ .window }}"}}])
		)
		/
		(sum(
			rate({{ .metricName }}{ {{ .additionalLabels }}{{ .serviceLabelName }}=~"{{ .serviceLabelValue }}"}[{{"{{ .window }}"}}])
		) > 0)
	) AND on({{ .serviceLabelName }}) sum(rate({{ .metricName }}{ {{ .serviceLabelName }}=~"{{ .serviceLabelValue }}"}[{{"{{ .window }}"}}])) > {{ .minimumAmountOfTraffic }}
) OR on() vector(0)
`))

// SLIPlugin will return a query that will return the availability error based on traefik V1 service metrics.
func SLIPlugin(ctx context.Context, meta, labels, options map[string]string) (string, error) {
	metricName, err := getMetricName(options)
	serviceLabelName, err := getServiceLabelName(options)
	serviceLabelValue, err := getServiceLabelValue(options)
	errorLabelName, err := getErrorLabelName(options)
	errorLabelValue, err := getErrorLabelValue(options)
	minimumAmountOfTraffic, err := getMinimumAmountOfTraffic(options)

	if err != nil {
		return "", fmt.Errorf("Error parsing options: %w", err)
	}

	var b bytes.Buffer
	data := map[string]string{
		"metricName":             metricName,
		"serviceLabelName":       serviceLabelName,
		"serviceLabelValue":      serviceLabelValue,
		"errorLabelName":         errorLabelName,
		"errorLabelValue":        errorLabelValue,
		"additionalLabels":       getAdditionalLabels(options),
		"minimumAmountOfTraffic": minimumAmountOfTraffic,
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

func getServiceLabelName(options map[string]string) (string, error) {
	label := options["serviceLabelName"]
	label = strings.TrimSpace(label)

	if label == "" {
		return "", fmt.Errorf("'serviceLabelName' name is required")
	}

	return label, nil
}

func getServiceLabelValue(options map[string]string) (string, error) {
	value := options["serviceLabelValue"]
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf("'serviceLabelValue' is required")
	}

	_, err := regexp.Compile(value)
	if err != nil {
		return "", fmt.Errorf("invalid regex for 'serviceLabelValue': %w", err)
	}

	return value, nil
}

func getErrorLabelName(options map[string]string) (string, error) {
	label := options["errorLabelName"]
	label = strings.TrimSpace(label)

	if label == "" {
		return "", fmt.Errorf("'errorLabelName' name is required")
	}

	return label, nil
}

func getErrorLabelValue(options map[string]string) (string, error) {
	value := options["errorLabelValue"]
	value = strings.TrimSpace(value)

	if value == "" {
		return "", fmt.Errorf("'errorLabelValue' is required")
	}

	_, err := regexp.Compile(value)
	if err != nil {
		return "", fmt.Errorf("invalid regex for 'errorLabelValue': %w", err)
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

func getMinimumAmountOfTraffic(options map[string]string) (string, error) {
	minimumAmountOfTraffic := options["minimumAmountOfTraffic"]
	if minimumAmountOfTraffic == "" {
		return "", fmt.Errorf("'minimumAmountOfTraffic' is required")
	}

	return minimumAmountOfTraffic, nil
}
