// Package metrics defines data structures and helper functions for working with
// application performance metrics of types "gauge" and "counter".
package metrics

import (
	"errors"
	"strconv"
)

var (
	// ErrIncorrectMetricValue is returned when a metric value is invalid or cannot be parsed.
	ErrIncorrectMetricValue = errors.New("incorrect value")
	// ErrIncorrectMetricType is returned when a metric type is not recognized.
	ErrIncorrectMetricType = errors.New("incorrect type")
	// ErrUnallowedMetric is returned when a metric is not allowed or not found.
	ErrUnallowedMetric = errors.New("unallowed metric")
)

// Gauge represents a metric of type "gauge" (floating-point value).
type Gauge float64

// Counter represents a metric of type "counter" (integer value).
type Counter int64

const (
	// GaugeName is the string identifier for the gauge metric type.
	GaugeName = "gauge"
	// CounterName is the string identifier for the counter metric type.
	CounterName = "counter"
)

func (val Gauge) String() string {
	s := strconv.FormatFloat(float64(val), 'f', -1, 64)
	if s == strconv.Itoa(int(val)) {
		s += "."
	}
	return s
}

func (val Counter) String() string {
	return strconv.FormatInt(int64(val), 10)
}

// RawMetric represents the transport format of a metric.
//
// It includes a metric ID, type, and optional value (gauge) or delta (counter).
//
//easyjson:json
type RawMetric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // counter
	Value *float64 `json:"value,omitempty"` // gauge
}

// RawMetrics is a slice of RawMetric.
//
//easyjson:json
type RawMetrics []RawMetric

// Metric represents an internal representation of a metric.
//
// Only one of Value (for gauge) or Delta (for counter) should be set.
type Metric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"` // counter
	Value float64 `json:"value,omitempty"` // gauge
}

// Metrics is a slice of Metric.
//
//easyjson:json
type Metrics []Metric

func NewMetric(id string, mType string, val string) (Metric, error) {
	switch mType {
	case GaugeName:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return Metric{}, ErrIncorrectMetricValue
		}
		return Metric{
			ID:    id,
			MType: mType,
			Value: float64(v),
		}, nil
	case CounterName:
		v, err := strconv.Atoi(val)
		if err != nil {
			return Metric{}, ErrIncorrectMetricValue
		}
		return Metric{
			ID:    id,
			MType: mType,
			Delta: int64(v),
		}, nil
	default:
		return Metric{}, ErrIncorrectMetricType
	}
}

// IsTypeAllowed returns true if the given metric type is supported (gauge or counter).
func IsTypeAllowed(mType string) bool {
	return mType == GaugeName || mType == CounterName
}

func (m Metric) StrValue() string {
	switch m.MType {
	case GaugeName:
		return Gauge(m.Value).String()
	case CounterName:
		return Counter(m.Delta).String()
	default:
		return "none"
	}
}

// Serialize converts a RawMetric to a Metric.
func (m RawMetric) Serialize() Metric {
	if m.Delta != nil {
		return Metric{
			ID:    m.ID,
			MType: m.MType,
			Delta: *m.Delta,
		}
	} else if m.Value != nil {
		return Metric{
			ID:    m.ID,
			MType: m.MType,
			Value: *m.Value,
		}
	} else {
		return Metric{
			ID:    m.ID,
			MType: m.MType,
		}
	}
}

// Serialize converts a RawMetrics slice to a Metrics slice.
func (m RawMetrics) Serialize() Metrics {
	result := make([]Metric, len(m))
	for i := range m {
		result[i].ID = m[i].ID
		result[i].MType = m[i].MType

		if m[i].Delta != nil {
			result[i].Delta = *m[i].Delta
		} else if m[i].Value != nil {
			result[i].Value = *m[i].Value
		}
	}
	return Metrics(result)
}

// Deserialize converts a Metric to a RawMetric.
func (m Metric) Deserialize() RawMetric {
	switch m.MType {
	case GaugeName:
		return RawMetric{
			ID:    m.ID,
			MType: m.MType,
			Value: &m.Value,
		}
	case CounterName:
		return RawMetric{
			ID:    m.ID,
			MType: m.MType,
			Delta: &m.Delta,
		}
	default:
		return RawMetric{
			ID:    m.ID,
			MType: m.MType,
		}
	}
}

// Deserialize converts a Metrics slice to a RawMetrics slice.
func (m Metrics) Deserialize() RawMetrics {
	result := make([]RawMetric, len(m))
	for i := range m {
		switch m[i].MType {
		case GaugeName:
			result[i] = RawMetric{
				ID:    m[i].ID,
				MType: m[i].MType,
				Value: &m[i].Value,
			}
		case CounterName:
			result[i] = RawMetric{
				ID:    m[i].ID,
				MType: m[i].MType,
				Delta: &m[i].Delta,
			}
		default:
			result[i] = RawMetric{
				ID:    m[i].ID,
				MType: m[i].MType,
			}
		}
	}
	return RawMetrics(result)
}
