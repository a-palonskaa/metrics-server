package metrics

import (
	"errors"
	"strconv"
)

var (
	ErrIncorrectMetricValue = errors.New("incorrect value")
	ErrIncorrectMetricType  = errors.New("incorrect type")
	ErrUnallowedMetric      = errors.New("unallowed metric")
)

type Gauge float64
type Counter int64

const (
	GaugeName   = "gauge"
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

//easyjson:json
type RawMetric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // counter
	Value *float64 `json:"value,omitempty"` // gauge
}

//easyjson:json
type RawMetrics []RawMetric

type Metric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"` // counter
	Value float64 `json:"value,omitempty"` // gauge
}

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
