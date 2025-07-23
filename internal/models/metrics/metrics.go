package metrics

import (
	"strconv"
)

type Gauge float64
type Counter int64

const GaugeName = "gauge"
const CounterName = "counter"

func (val Gauge) String() string {
	return strconv.FormatFloat(float64(val), 'f', 1, 64)
}

func (val Counter) String() string {
	return strconv.FormatInt(int64(val), 10)
}

//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // counter
	Value *float64 `json:"value,omitempty"` // gauge
}

//easyjson:json
type Metrics []Metric

func NewMetrics(id string, mType string, val string) (Metric, error) {
	switch mType {
	case GaugeName:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return Metric{}, ErrIncorrectMetricValue
		}
		gauge := float64(v)
		return Metric{
			ID:    id,
			MType: mType,
			Value: &gauge,
		}, nil
	case CounterName:
		v, err := strconv.Atoi(val)
		if err != nil {
			return Metric{}, ErrIncorrectMetricValue
		}
		counter := int64(v)
		return Metric{
			ID:    id,
			MType: mType,
			Delta: &counter,
		}, nil
	default:
		return Metric{}, ErrIncorrectMetricType
	}
}

func IsTypeAllowed(mType string) bool {
	return mType == GaugeName || mType == CounterName
}

func (m Metric) StrValue() string {
	if m.Delta != nil {
		return Counter(*m.Delta).String()
	}
	return Gauge(*m.Value).String()
}
