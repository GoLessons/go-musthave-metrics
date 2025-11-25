package convert

import (
	"fmt"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/proto"
)

func ProtoToModel(protoMetric *proto.Metric) (model.Metrics, error) {
	var result model.Metrics
	result.ID = protoMetric.Id
	switch protoMetric.Type {
	case proto.Metric_GAUGE:
		result.MType = model.Gauge
		value := protoMetric.Value
		result.Value = &value
	case proto.Metric_COUNTER:
		result.MType = model.Counter
		delta := protoMetric.Delta
		result.Delta = &delta
	default:
		return result, fmt.Errorf("unknown metric type")
	}
	return result, nil
}

func ModelToProto(metric model.Metrics) (*proto.Metric, error) {
	result := &proto.Metric{Id: metric.ID}
	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return nil, fmt.Errorf("missing value")
		}
		result.Type = proto.Metric_GAUGE
		result.Value = *metric.Value
	case model.Counter:
		if metric.Delta == nil {
			return nil, fmt.Errorf("missing delta")
		}
		result.Type = proto.Metric_COUNTER
		result.Delta = *metric.Delta
	default:
		return nil, fmt.Errorf("unknown metric type")
	}
	return result, nil
}
