package custommetrics

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
)

type otelCollectorMetricProvider struct {
}

func (o *otelCollectorMetricProvider) GetMetricByName(ctx context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	logger := log.FromContext(ctx)
	logger.Info("GetMetricByName", "name", name, "info", info, "metricSelector", metricSelector)
	return &custom_metrics.MetricValue{}, nil
}

func (o *otelCollectorMetricProvider) GetMetricBySelector(ctx context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	logger := log.FromContext(ctx)
	logger.Info("GetMetricBySelector", "namespace", namespace, "selector", selector, "info", info, "metricSelector", metricSelector)
	return &custom_metrics.MetricValueList{}, nil
}

func (o *otelCollectorMetricProvider) ListAllMetrics() []provider.CustomMetricInfo {
	return []provider.CustomMetricInfo{
		{
			Metric: "collector-lost-spans",
		},
	}
}
