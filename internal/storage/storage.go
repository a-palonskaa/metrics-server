package memstorage

type Gauge float64
type Counter int64

type MemStorage struct {
	Gauge   map[string]Gauge
	Counter map[string]Counter
}

type MetricsStorage interface {
	AddGauge(string, Gauge)
	AddCounter(string, Counter)
}

func (ms *MemStorage) AddGauge(name string, val Gauge) {
	ms.Gauge[name] = val
}

func (ms *MemStorage) AddCounter(name string, val Counter) {
	ms.Counter[name] += val
}

var MS = MemStorage{
	Gauge:   make(map[string]Gauge),
	Counter: make(map[string]Counter),
}
