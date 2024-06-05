package availability_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	errorrate "github.com/lokalise/common-sloth-sli-plugins/plugins/http-error-rate"
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

		"An invalid service label name, should fail.": {
			options: map[string]string{"serviceLabelName": "([xyz"},
			expErr:  true,
		},

		"An invalid service label value, should fail.": {
			options: map[string]string{"serviceLabelName": "([xyz"},
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

		"When all options are provided, it should return a valid query.": {
			options: map[string]string{
				"metricName":        "http_request_duration_seconds_count",
				"serviceLabelName":  "service",
				"serviceLabelValue": "test",
				"errorLabelName":    "status_code",
				"errorLabelValue":   "(5..|429|431)",
				"additionalLabels":  "route=~\".*\"",
			},
			expQuery: `
(
	sum(
		rate(http_request_duration_seconds_count{ route=~".*", service=~"test", status_code=~"(5..|429|431)"}[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ route=~".*", service=~"test"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotQuery, err := errorrate.SLIPlugin(context.TODO(), test.meta, test.labels, test.options)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expQuery, gotQuery)
			}
		})
	}
}
