package collectorconfig

import (
	"fmt"

	"github.com/odigos-io/odigos/api/k8sconsts"
	"github.com/odigos-io/odigos/common/config"
)

const (
	ebpfProfilerReceiverName             = "profiling"
	profilesPipelineName                 = "profiles"
	clusterCollectorProfilesExporterName = "otlp/out-cluster-collector-profiles"
	k8sAttributesProcessorName           = "k8sattributes"
)

func ProfilesConfig(odigosNamespace string) config.Config {
	return config.Config{
		Receivers: config.GenericMap{
			ebpfProfilerReceiverName: config.GenericMap{
				// Use defaults for production - the profiler will auto-enable
				// all supported tracers (perl,php,python,hotspot,ruby,v8,go)
			},
		},
		Processors: config.GenericMap{
			k8sAttributesProcessorName: config.GenericMap{
				"passthrough": false,
				"pod_association": []config.GenericMap{
					{
						"sources": []config.GenericMap{
							{
								"from": "resource_attribute",
								"name": "k8s.pod.uid",
							},
						},
					},
				},
				"extract": config.GenericMap{
					"metadata": []string{
						"k8s.namespace.name",
						"k8s.pod.name",
						"k8s.pod.uid",
						"k8s.deployment.name",
						"k8s.daemonset.name",
						"k8s.statefulset.name",
						"k8s.replicaset.name",
						"k8s.container.name",
						"k8s.node.name",
					},
					"labels": []config.GenericMap{
						{
							"key_regex": "odigos\\.io/.*",
							"from":      "pod",
						},
					},
				},
				"filter": config.GenericMap{
					"node_from_env_var": "NODE_NAME",
				},
			},
		},
		Exporters: config.GenericMap{
			clusterCollectorProfilesExporterName: config.GenericMap{
				"endpoint": fmt.Sprintf("dns:///%s.%s:4317", k8sconsts.OdigosClusterCollectorDeploymentName, odigosNamespace),
				"tls": config.GenericMap{
					"insecure": true,
				},
			},
		},
		Service: config.Service{
			Pipelines: map[string]config.Pipeline{
				profilesPipelineName: {
					Receivers:  []string{ebpfProfilerReceiverName},
					Processors: []string{k8sAttributesProcessorName},
					Exporters:  []string{clusterCollectorProfilesExporterName},
				},
			},
		},
	}
}
