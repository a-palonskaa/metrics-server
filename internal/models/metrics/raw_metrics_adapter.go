package metrics

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
