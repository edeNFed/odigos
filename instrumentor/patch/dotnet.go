package patch

import (
	"fmt"
	odigosv1 "github.com/keyval-dev/odigos/api/v1alpha1"
	"github.com/keyval-dev/odigos/common"
	"github.com/keyval-dev/odigos/common/consts"
	v1 "k8s.io/api/core/v1"
)

const (
	dotnetAgentName          = "keyval/dotnet-agent:v0.1"
	enableProfilingEnvVar    = "CORECLR_ENABLE_PROFILING"
	profilerEndVar           = "CORECLR_PROFILER"
	profilerId               = "{918728DD-259F-4A6A-AC2B-B85E1B658318}"
	profilerPathEnv          = "CORECLR_PROFILER_PATH"
	profilerPath             = "/agent/OpenTelemetry.AutoInstrumentation.Native.so"
	dotNetStartupHookEnv     = "DOTNET_STARTUP_HOOKS"
	dotNetStartupHookPath    = "/agent/netcoreapp3.1/OpenTelemetry.AutoInstrumentation.StartupHook.dll"
	dotNetAdditionalDeps     = "DOTNET_ADDITIONAL_DEPS"
	serviceNameEnv           = "OTEL_SERVICE_NAME"
	dotNetAdditionalDepsPath = "/agent/AdditionalDeps"
	collectorUrlEnv          = "OTEL_EXPORTER_OTLP_ENDPOINT"
	dotNetSharedStore        = "DOTNET_SHARED_STORE"
	dotNetSharedStorePath    = "/agent/store"
	tracerHomeEnv            = "OTEL_DOTNET_AUTO_HOME"
	exportTypeEnv            = "OTEL_TRACES_EXPORTER"
	tracerHome               = "/agent"
	dotnetVolumeName         = "agentdir-dotnet"
	dotNetPropagatorEnv      = "OTEL_PROPAGATORS"
	dotNetPropagatorEnvVal   = "tracecontext"
)

var dotNet = &dotNetPatcher{}

type dotNetPatcher struct{}

func (d *dotNetPatcher) Patch(podSpec *v1.PodTemplateSpec, instrumentation *odigosv1.InstrumentedApplication) {
	podSpec.Spec.Volumes = append(podSpec.Spec.Volumes, v1.Volume{
		Name: dotnetVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	podSpec.Spec.InitContainers = append(podSpec.Spec.InitContainers, v1.Container{
		Name:    "copy-dotnet-agent",
		Image:   dotnetAgentName,
		Command: []string{"cp", "-a", "/autoinstrumentation/.", "/agent/"},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      dotnetVolumeName,
				MountPath: tracerHome,
			},
		},
	})

	var modifiedContainers []v1.Container
	for _, container := range podSpec.Spec.Containers {
		if shouldPatch(instrumentation, common.DotNetProgrammingLanguage, container.Name) {
			container.Env = append([]v1.EnvVar{{
				Name: NodeIPEnvName,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			}}, container.Env...)

			container.Env = append(container.Env, v1.EnvVar{
				Name:  enableProfilingEnvVar,
				Value: "1",
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  profilerEndVar,
				Value: profilerId,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  profilerPathEnv,
				Value: profilerPath,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  dotNetStartupHookEnv,
				Value: dotNetStartupHookPath,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  dotNetAdditionalDeps,
				Value: dotNetAdditionalDepsPath,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  tracerHomeEnv,
				Value: tracerHome,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  dotNetSharedStore,
				Value: dotNetSharedStorePath,
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  dotNetPropagatorEnv,
				Value: dotNetPropagatorEnvVal,
			})

			// Currently .NET instrumentation only support zipkin format, we should move to OTLP when support is added
			container.Env = append(container.Env, v1.EnvVar{
				Name:  collectorUrlEnv,
				Value: fmt.Sprintf("http://%s:%d", HostIPEnvValue, consts.OTLPHttpPort),
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  serviceNameEnv,
				Value: calculateAppName(podSpec, &container, instrumentation),
			})

			container.Env = append(container.Env, v1.EnvVar{
				Name:  exportTypeEnv,
				Value: "otlp",
			})

			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				MountPath: tracerHome,
				Name:      dotnetVolumeName,
			})
		}

		modifiedContainers = append(modifiedContainers, container)
	}

	podSpec.Spec.Containers = modifiedContainers
}

func (d *dotNetPatcher) IsInstrumented(podSpec *v1.PodTemplateSpec, instrumentation *odigosv1.InstrumentedApplication) bool {
	// TODO: Deep comparison
	for _, c := range podSpec.Spec.InitContainers {
		if c.Name == "copy-dotnet-agent" {
			return true
		}
	}
	return false
}
