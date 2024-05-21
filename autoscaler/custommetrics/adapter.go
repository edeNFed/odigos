package custommetrics

import (
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
)

type Adapter struct {
	basecmd.AdapterBase
}

func (a *Adapter) NewProvider() (provider.CustomMetricsProvider, error) {
	return &otelCollectorMetricProvider{}, nil
}
