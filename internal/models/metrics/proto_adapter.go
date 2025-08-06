package metrics

import (
	proto "github.com/a-palonskaa/metrics-server/gen/proto"
)

// ProtoToModel converts a Proto Metric structure to a Metric.
func ProtoToModel(pm *proto.Metric) Metric {
	mt := Metric{
		ID:    pm.Id,
		MType: pm.MType,
	}

	switch pm.MType {
	case CounterName:
		if v, ok := pm.MetricValue.(*proto.Metric_Delta); ok {
			mt.Delta = v.Delta
		}
	case GaugeName:
		if v, ok := pm.MetricValue.(*proto.Metric_Value); ok {
			mt.Value = v.Value
		}
	}
	return mt
}

// ModelToProto converts a Metric to a Proto Metric structure.
func ModelToProto(m Metric) *proto.Metric {
	pb := &proto.Metric{
		Id:    m.ID,
		MType: m.MType,
	}

	switch m.MType {
	case CounterName:
		pb.MetricValue = &proto.Metric_Delta{Delta: m.Delta}
	case GaugeName:
		pb.MetricValue = &proto.Metric_Value{Value: m.Value}
	}
	return pb
}

// ProtoListToModel converts a Proto Metric List to a Metrics.
func ProtoListToModel(pbMetrics []*proto.Metric) Metrics {
	result := make([]Metric, 0, len(pbMetrics))
	for _, pb := range pbMetrics {
		result = append(result, ProtoToModel(pb))
	}
	return Metrics(result)
}

// ModelListToProto converts a Metrics to a Proto Metric List.
func ModelListToProto(metricsList Metrics) []*proto.Metric {
	result := make([]*proto.Metric, 0, len(metricsList))

	for _, m := range metricsList {
		result = append(result, ModelToProto(m))
	}
	return result
}
