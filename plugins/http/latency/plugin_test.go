package availability_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	latency "github.com/lokalise/common-sloth-sli-plugins/plugins/http/latency"
)

func TestSLIPlugin(t *testing.T) {
	tests := map[string]struct {
		meta     map[string]string
		labels   map[string]string
		options  map[string]string
		expQuery string
		expErr   bool
	}{
		"Without service name, should fail.": {
			options: map[string]string{},
			expErr:  true,
		},

		"An invalid service name query, should fail.": {
			options: map[string]string{"service_name_regex": "([xyz"},
			expErr:  true,
		},

		"An empty service name query, should fail.": {
			options: map[string]string{"service_name_regex": ""},
			expErr:  true,
		},

		"Not having a filter and with service name should return a valid query.": {
			options: map[string]string{
				"service_name_regex": "test",
				"bucket":             "0.5",
			},
			expQuery: `
1 - (
	sum(
		rate(http_request_duration_seconds_bucket{ service=~"test", route=~".*", le="0.5" }[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ service=~"test", route=~".*"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},

		"Having a filter and with service name should return a valid query.": {
			options: map[string]string{
				"filter":             `k1="v2",k2="v2"`,
				"bucket":             "0.5",
				"service_name_regex": "test",
			},
			expQuery: `
1 - (
	sum(
		rate(http_request_duration_seconds_bucket{ k1="v2",k2="v2",service=~"test", route=~".*", le="0.5" }[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ k1="v2",k2="v2",service=~"test", route=~".*"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},

		"Filter should be sanitized with ','.": {
			options: map[string]string{
				"filter":             `k1="v2",k2="v2",`,
				"bucket":             "0.5",
				"service_name_regex": "test",
			},
			expQuery: `
1 - (
	sum(
		rate(http_request_duration_seconds_bucket{ k1="v2",k2="v2",service=~"test", route=~".*", le="0.5" }[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ k1="v2",k2="v2",service=~"test", route=~".*"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},

		"Filter should be sanitized with '{'.": {
			options: map[string]string{
				"filter":             `{k1="v2",k2="v2",},`,
				"bucket":             "0.5",
				"service_name_regex": "test",
			},
			expQuery: `
1 - (
	sum(
		rate(http_request_duration_seconds_bucket{ k1="v2",k2="v2",service=~"test", route=~".*", le="0.5" }[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ k1="v2",k2="v2",service=~"test", route=~".*"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},
		"Route provided": {
			options: map[string]string{
				"filter":             `{k1="v2",k2="v2",},`,
				"service_name_regex": "test",
				"bucket":             "0.5",
				"route_regex":        "/test.+",
			},
			expQuery: `
1 - (
	sum(
		rate(http_request_duration_seconds_bucket{ k1="v2",k2="v2",service=~"test", route=~"/test.+", le="0.5" }[{{ .window }}])
	)
	/
	(sum(
		rate(http_request_duration_seconds_count{ k1="v2",k2="v2",service=~"test", route=~"/test.+"}[{{ .window }}])
	) > 0)
) OR on() vector(0)
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotQuery, err := latency.SLIPlugin(context.TODO(), test.meta, test.labels, test.options)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expQuery, gotQuery)
			}
		})
	}
}
