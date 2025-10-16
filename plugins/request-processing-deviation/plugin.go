package requestprocessingdeviation

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
	SLIPluginID      = "lokalise/request-processing-deviation"
)

var queryTpl = template.Must(template.New("").Option("missingkey=error").Parse(`
(
	(
		(
			sum(
				rate({{ .metricName }}{ {{ .additionalLabels }}{{ .serviceLabelName }}=~"{{ .serviceLabelValue }}", {{ .statusLabelName }}="{{ .requestedStatus }}"}[{{"{{ .window }}"}}])
			)
			-
			sum(
				rate({{ .metricName }}{ {{ .additionalLabels }}{{ .serviceLabelName }}=~"{{ .serviceLabelValue }}", {{ .statusLabelName }}="{{ .processedStatus }}"}[{{"{{ .window }}"}}])
			)
		)
		/
		(sum(
			rate({{ .metricName }}{ {{ .additionalLabels }}{{ .serviceLabelName }}=~"{{ .serviceLabelValue }}", {{ .statusLabelName }}="{{ .requestedStatus }}"}[{{"{{ .window }}"}}])
		) > 0)
	){{ if ne .minimumRequestsPerSecond "0" }} AND on() sum(rate({{ .metricName }}{ {{ .serviceLabelName }}=~"{{ .serviceLabelValue }}"}[{{"{{ .window }}"}}])) > {{ .minimumRequestsPerSecond }}{{ end }}
) OR on() vector(0)
`))

// SLIPlugin will return a query that calculates the deviation between requested and processed counters.
func SLIPlugin(ctx context.Context, meta, labels, options map[string]string) (string, error) {
	metricName, err := getMetricName(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	serviceLabelName, err := getServiceLabelName(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	serviceLabelValue, err := getServiceLabelValue(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	statusLabelName, err := getStatusLabelName(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	requestedStatus, err := getRequestedStatus(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	processedStatus, err := getProcessedStatus(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	minimumRequestsPerSecond, err := getMinimumRequestsPerSecond(options)
	if err != nil {
		return "", fmt.Errorf("error parsing options: %w", err)
	}

	var b bytes.Buffer
	data := map[string]string{
		"metricName":               metricName,
		"serviceLabelName":         serviceLabelName,
		"serviceLabelValue":        serviceLabelValue,
		"statusLabelName":          statusLabelName,
		"requestedStatus":          requestedStatus,
		"processedStatus":          processedStatus,
		"additionalLabels":         getAdditionalLabels(options),
		"minimumRequestsPerSecond": minimumRequestsPerSecond,
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

func getStatusLabelName(options map[string]string) (string, error) {
	label := options["statusLabelName"]
	label = strings.TrimSpace(label)

	if label == "" {
		return "", fmt.Errorf("'statusLabelName' is required")
	}

	return label, nil
}

func getRequestedStatus(options map[string]string) (string, error) {
	status := options["requestedStatus"]
	status = strings.TrimSpace(status)

	if status == "" {
		return "", fmt.Errorf("'requestedStatus' is required")
	}

	return status, nil
}

func getProcessedStatus(options map[string]string) (string, error) {
	status := options["processedStatus"]
	status = strings.TrimSpace(status)

	if status == "" {
		return "", fmt.Errorf("'processedStatus' is required")
	}

	return status, nil
}

func getMetricName(options map[string]string) (string, error) {
	metricName := options["metricName"]
	if metricName == "" {
		return "", fmt.Errorf("'metricName' is required")
	}

	return metricName, nil
}

func getMinimumRequestsPerSecond(options map[string]string) (string, error) {
	minimumRequestsPerSecond := options["minimumRequestsPerSecond"]
	if minimumRequestsPerSecond == "" {
		// Default to 0 if not provided, meaning no minimum threshold
		return "0", nil
	}

	return minimumRequestsPerSecond, nil
}
