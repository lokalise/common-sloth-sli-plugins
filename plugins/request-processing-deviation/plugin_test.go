package requestprocessingdeviation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	requestprocessingdeviation "github.com/lokalise/common-sloth-sli-plugins/plugins/request-processing-deviation"
)

func TestSLIPlugin(t *testing.T) {
	tests := map[string]struct {
		meta     map[string]string
		labels   map[string]string
		options  map[string]string
		expQuery string
		expErr   bool
	}{
		"Without anything provided, should fail.": {
			options: map[string]string{},
			expErr:  true,
		},

		"Empty service label name, should fail.": {
			options: map[string]string{"serviceLabelName": ""},
			expErr:  true,
		},

		"Empty metric name, should fail.": {
			options: map[string]string{"metricName": ""},
			expErr:  true,
		},

		"Empty service label value, should fail.": {
			options: map[string]string{"serviceLabelValue": ""},
			expErr:  true,
		},

		"Empty status label name, should fail.": {
			options: map[string]string{"statusLabelName": ""},
			expErr:  true,
		},

		"Empty requested status, should fail.": {
			options: map[string]string{"requestedStatus": ""},
			expErr:  true,
		},

		"Empty processed status, should fail.": {
			options: map[string]string{"processedStatus": ""},
			expErr:  true,
		},

		"When all required options are provided without minimumRequestsPerSecond, it should return a valid query.": {
			options: map[string]string{
				"metricName":        "request_counter_total",
				"serviceLabelName":  "service",
				"serviceLabelValue": "test",
				"statusLabelName":   "status",
				"requestedStatus":   "requested",
				"processedStatus":   "processed",
				"additionalLabels":  "route=~\".*\"",
			},
			expQuery: `
(
	(
		(
			sum(
				rate(request_counter_total{ route=~".*", service=~"test", status="requested"}[{{ .window }}])
			)
			-
			sum(
				rate(request_counter_total{ route=~".*", service=~"test", status="processed"}[{{ .window }}])
			)
		)
		/
		(sum(
			rate(request_counter_total{ route=~".*", service=~"test", status="requested"}[{{ .window }}])
		) > 0)
	)
) OR on() vector(0)
`,
		},

		"When all options are provided with minimumRequestsPerSecond, it should return a valid query.": {
			options: map[string]string{
				"metricName":               "request_counter_total",
				"serviceLabelName":         "service",
				"serviceLabelValue":        "test",
				"statusLabelName":          "status",
				"requestedStatus":          "requested",
				"processedStatus":          "processed",
				"additionalLabels":         "route=~\".*\"",
				"minimumRequestsPerSecond": "10",
			},
			expQuery: `
(
	(
		(
			sum(
				rate(request_counter_total{ route=~".*", service=~"test", status="requested"}[{{ .window }}])
			)
			-
			sum(
				rate(request_counter_total{ route=~".*", service=~"test", status="processed"}[{{ .window }}])
			)
		)
		/
		(sum(
			rate(request_counter_total{ route=~".*", service=~"test", status="requested"}[{{ .window }}])
		) > 0)
	) AND on() sum(rate(request_counter_total{ service=~"test"}[{{ .window }}])) > 10
) OR on() vector(0)
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assertions := assert.New(t)

			gotQuery, err := requestprocessingdeviation.SLIPlugin(context.TODO(), test.meta, test.labels, test.options)

			if test.expErr {
				assertions.Error(err)
			} else if assertions.NoError(err) {
				assertions.Equal(test.expQuery, gotQuery)
			}
		})
	}
}
