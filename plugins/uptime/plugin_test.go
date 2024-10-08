package uptime_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	uptime "github.com/lokalise/common-sloth-sli-plugins/plugins/uptime"
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

		"An invalid ingress label name, should fail.": {
			options: map[string]string{"ingressLabelName": "([xyz"},
			expErr:  true,
		},

		"An invalid ingress label value, should fail.": {
			options: map[string]string{"ingressLabelName": "([xyz"},
			expErr:  true,
		},

		"Empty ingress label name, should fail.": {
			options: map[string]string{"ingressLabelName": ""},
			expErr:  true,
		},

		"Empty metric name, should fail.": {
			options: map[string]string{"metricName": ""},
			expErr:  true,
		},

		"Empty ingress label value, should fail.": {
			options: map[string]string{"ingressLabelValue": ""},
			expErr:  true,
		},

		"When all options are provided, it should return a valid query.": {
			options: map[string]string{
				"metricName":               "up",
				"ingressLabelName":         "ingress",
				"ingressLabelValue":        "test",
				"additionalLabels":         "instance=~\".*\"",
			},
			expQuery: `
sum(count_over_time((up{ instance=~".*", ingress=~"test" } == 0)[{{ .window }}:])) or vector(0)
/
sum(count_over_time((up{ instance=~".*", ingress=~"test" })[{{ .window }}:]))
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotQuery, err := uptime.SLIPlugin(context.TODO(), test.meta, test.labels, test.options)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expQuery, gotQuery)
			}
		})
	}
}
