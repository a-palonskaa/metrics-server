package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

func TestGaugeString(t *testing.T) {
	tests := []struct {
		name     string
		input    metrics.Gauge
		expected string
	}{
		{name: "integer value",
			input:    metrics.Gauge(42),
			expected: "42.",
		},
		{name: "fractional value",
			input:    metrics.Gauge(3.1415),
			expected: "3.1415",
		},
		{name: "negative value",
			input:    metrics.Gauge(-10.5),
			expected: "-10.5",
		},
		{name: "zero value",
			input:    metrics.Gauge(0),
			expected: "0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}

func TestCounterString(t *testing.T) {
	tests := []struct {
		name     string
		input    metrics.Counter
		expected string
	}{
		{
			name:     "positive value",
			input:    metrics.Counter(100),
			expected: "100",
		},
		{
			name:     "zero value",
			input:    metrics.Counter(0),
			expected: "0",
		},
		{
			name:     "negative value",
			input:    metrics.Counter(-1),
			expected: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}

func TestStrValue(t *testing.T) {
	tests := []struct {
		name     string
		input    metrics.Metric
		expected string
	}{
		{
			name: "gauge",
			input: metrics.Metric{
				ID:    "temp",
				MType: metrics.GaugeName,
				Value: 36.6,
			},
			expected: "36.6",
		},
		{
			name: "counter",
			input: metrics.Metric{
				ID:    "req",
				MType: metrics.CounterName,
				Delta: 100,
			},
			expected: "100",
		},
		{
			name: "non",
			input: metrics.Metric{
				ID:    "req",
				MType: "meow",
			},
			expected: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.StrValue())
		})
	}
}

func TestIsTypeAllowed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "gauge-type",
			input:    metrics.GaugeName,
			expected: true,
		},
		{
			name:     "counter-name",
			input:    metrics.CounterName,
			expected: true,
		},
		{
			name:     "unknown",
			input:    "unknown",
			expected: false,
		},
		{
			name:     "empty",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, metrics.IsTypeAllowed(tt.input))
		})
	}
}

func TestSerializeDeserialize(t *testing.T) {
	t.Run("RawMetric to Metric", func(t *testing.T) {
		tests := []struct {
			name     string
			input    metrics.RawMetric
			expected metrics.Metric
		}{
			{
				name: "gauge metric",
				input: metrics.RawMetric{
					ID:    "temp",
					MType: metrics.GaugeName,
					Value: ptrFloat64(36.6),
				},
				expected: metrics.Metric{
					ID:    "temp",
					MType: metrics.GaugeName,
					Value: 36.6,
				},
			},
			{
				name: "counter metric",
				input: metrics.RawMetric{
					ID:    "req",
					MType: metrics.CounterName,
					Delta: ptrInt64(100),
				},
				expected: metrics.Metric{
					ID:    "req",
					MType: metrics.CounterName,
					Delta: 100,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.input.Serialize()
				assert.Equal(t, tt.expected, result)

				resultAfter := tt.expected.Deserialize()
				assert.Equal(t, resultAfter, tt.input)
			})
		}
	})

	t.Run("RawMetrics to Metrics", func(t *testing.T) {
		tests := []struct {
			name     string
			input    metrics.RawMetrics
			expected metrics.Metrics
		}{
			{
				name: "single gauge metric",
				input: metrics.RawMetrics{
					{
						ID:    "temp",
						MType: metrics.GaugeName,
						Value: ptrFloat64(36.6),
					},
				},
				expected: metrics.Metrics{
					{
						ID:    "temp",
						MType: metrics.GaugeName,
						Value: 36.6,
					},
				},
			},
			{
				name: "single counter metric",
				input: metrics.RawMetrics{
					{
						ID:    "requests",
						MType: metrics.CounterName,
						Delta: ptrInt64(100),
					},
				},
				expected: metrics.Metrics{
					{
						ID:    "requests",
						MType: metrics.CounterName,
						Delta: 100,
					},
				},
			},
			{
				name: "multiple metrics",
				input: metrics.RawMetrics{
					{
						ID:    "temp",
						MType: metrics.GaugeName,
						Value: ptrFloat64(36.6),
					},
					{
						ID:    "requests",
						MType: metrics.CounterName,
						Delta: ptrInt64(100),
					},
				},
				expected: metrics.Metrics{
					{
						ID:    "temp",
						MType: metrics.GaugeName,
						Value: 36.6,
					},
					{
						ID:    "requests",
						MType: metrics.CounterName,
						Delta: 100,
					},
				},
			},
			{
				name:     "empty metrics",
				input:    metrics.RawMetrics{},
				expected: metrics.Metrics{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.input.Serialize()
				assert.Equal(t, len(tt.expected), len(result))
				for i := range tt.expected {
					assert.Equal(t, tt.expected[i], result[i])
				}

				resultAfter := tt.expected.Deserialize()
				assert.Equal(t, len(tt.input), len(resultAfter))
				for i := range tt.input {
					assert.Equal(t, tt.input[i], resultAfter[i])
				}
			})
		}
	})
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt64(i int64) *int64 {
	return &i
}

func TestSerialize(t *testing.T) {
	t.Run("RawMetric to Metric", func(t *testing.T) {
		tests := []struct {
			name     string
			input    metrics.RawMetric
			expected metrics.Metric
		}{
			{
				name: "gauge metric",
				input: metrics.RawMetric{
					ID:    "temp",
					MType: metrics.GaugeName,
				},
				expected: metrics.Metric{
					ID:    "temp",
					MType: metrics.GaugeName,
				},
			},
			{
				name: "counter metric",
				input: metrics.RawMetric{
					ID:    "req",
					MType: metrics.CounterName,
				},
				expected: metrics.Metric{
					ID:    "req",
					MType: metrics.CounterName,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.input.Serialize()
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestDeserialize(t *testing.T) {
	t.Run("RawMetric to Metric", func(t *testing.T) {
		tests := []struct {
			name     string
			input    metrics.RawMetric
			expected metrics.Metric
		}{
			{
				name: "gauge metric",
				input: metrics.RawMetric{
					ID:    "temp",
					MType: "meow",
				},
				expected: metrics.Metric{
					ID:    "temp",
					MType: "meow",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resultAfter := tt.expected.Deserialize()
				assert.Equal(t, resultAfter, tt.input)
			})
		}
	})
}
