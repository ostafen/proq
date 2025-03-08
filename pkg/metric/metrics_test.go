package metric

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMetricLine(t *testing.T) {
	type testCase struct {
		line           string
		expectedMetric RawMetric
		expectedErr    error
	}

	cases := []testCase{
		{
			line: `http_requests_total{method="get", status="200"} 12345`,
			expectedMetric: RawMetric{
				Name: "http_requests_total",
				Labels: []Label{
					{
						Name:  "method",
						Value: "get",
					},
					{
						Name:  "status",
						Value: "200",
					},
				},
				Value: 12345,
			},
			expectedErr: nil,
		},
		{
			line: `http_requests_total{method="post", status="404", endpoint="/api/v1"} 98765`,
			expectedMetric: RawMetric{
				Name: "http_requests_total",
				Labels: []Label{
					{
						Name:  "method",
						Value: "post",
					},
					{
						Name:  "status",
						Value: "404",
					},
					{
						Name:  "endpoint",
						Value: "/api/v1",
					},
				},
				Value: 98765,
			},
			expectedErr: nil,
		},
		{
			line: `http_requests_total 1234`,
			expectedMetric: RawMetric{
				Name:   "http_requests_total",
				Labels: nil,
				Value:  1234,
			},
			expectedErr: nil,
		},
		{
			line: `memory_usage{unit="MB"} 1234.56`,
			expectedMetric: RawMetric{
				Name: "memory_usage",
				Labels: []Label{
					{
						Name:  "unit",
						Value: "MB",
					},
				},
				Value: 1234.56,
			},
			expectedErr: nil,
		},
		{
			line: `disk_io{device="sda", direction="write"} 1000000`,
			expectedMetric: RawMetric{
				Name: "disk_io",
				Labels: []Label{
					{
						Name:  "device",
						Value: "sda",
					},
					{
						Name:  "direction",
						Value: "write",
					},
				},
				Value: 1000000,
			},
			expectedErr: nil,
		},
		{
			line:           `http_requests_total{method="get", status="200"}`,
			expectedMetric: RawMetric{},
			expectedErr:    ErrInvalidMetricLine,
		},
		{
			line:           ``,
			expectedMetric: RawMetric{},
			expectedErr:    ErrInvalidMetricLine,
		},
		{
			line: `requests_total{method="post", status="200", user="alice@xyz.com"} 500`,
			expectedMetric: RawMetric{
				Name: "requests_total",
				Labels: []Label{
					{
						Name:  "method",
						Value: "post",
					},
					{
						Name:  "status",
						Value: "200",
					},
					{
						Name:  "user",
						Value: "alice@xyz.com",
					},
				},
				Value: 500,
			},
			expectedErr: nil,
		},
		{
			line: `requests_total{method="get", status="500"}   123`,
			expectedMetric: RawMetric{
				Name: "requests_total",
				Labels: []Label{
					{
						Name:  "method",
						Value: "get",
					},
					{
						Name:  "status",
						Value: "500",
					},
				},
				Value: 123,
			},
			expectedErr: nil,
		},
		{
			line: `cpu_usage 75.5`,
			expectedMetric: RawMetric{
				Name:   "cpu_usage",
				Labels: nil,
				Value:  75.5,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.line, func(t *testing.T) {
			m, err := ParseMetricLine(tc.line)
			if tc.expectedErr != nil {
				require.Zero(t, m)
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.Equal(t, tc.expectedMetric, m)
				require.NoError(t, err)
			}
		})
	}
}
