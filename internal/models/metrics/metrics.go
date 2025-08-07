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
