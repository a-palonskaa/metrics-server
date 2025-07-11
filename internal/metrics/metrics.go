package metrics

import (
	"strconv"
)

type Gauge float64
type Counter int64

const GaugeName = "gauge"
const CounterName = "counter"

func (val Gauge) String() string {
	return strconv.FormatFloat(float64(val), 'f', -1, 64)
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
