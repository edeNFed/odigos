package workload

import (
	"context"
	"errors"
	"strings"

	"github.com/odigos-io/odigos/common"

	"github.com/odigos-io/odigos/common/consts"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Workload interface {
	client.Object
	AvailableReplicas() int32
}

// compile time check for interface implementation
var _ Workload = &DeploymentWorkload{}
var _ Workload = &DaemonSetWorkload{}
var _ Workload = &StatefulSetWorkload{}

type DeploymentWorkload struct {
	*v1.Deployment
}

func (d *DeploymentWorkload) AvailableReplicas() int32 {
	return d.Status.AvailableReplicas
}

type DaemonSetWorkload struct {
	*v1.DaemonSet
}

func (d *DaemonSetWorkload) AvailableReplicas() int32 {
	return d.Status.NumberReady
}

type StatefulSetWorkload struct {
	*v1.StatefulSet
}

func (s *StatefulSetWorkload) AvailableReplicas() int32 {
	return s.Status.ReadyReplicas
}

func ObjectToWorkload(obj client.Object) (Workload, error) {
	switch t := obj.(type) {
	case *v1.Deployment:
		return &DeploymentWorkload{Deployment: t}, nil
	case *v1.DaemonSet:
		return &DaemonSetWorkload{DaemonSet: t}, nil
	case *v1.StatefulSet:
		return &StatefulSetWorkload{StatefulSet: t}, nil
	default:
		return nil, errors.New("unknown kind")
	}
}

func IsContainerInstrumented(c *corev1.Container) bool {
	if c != nil && c.Resources.Limits != nil {
		for val := range c.Resources.Limits {
			if strings.HasPrefix(val.String(), common.OdigosResourceNamespace) {
				return true
			}
		}
	}

	return false
}

func IsObjectLabeledForInstrumentation(obj client.Object) bool {
	labels := obj.GetLabels()
	if labels == nil {
		return false
	}

	val, exists := labels[consts.OdigosInstrumentationLabel]
	if !exists {
		return false
	}

	return val == consts.InstrumentationEnabled
}

func IsInstrumentationDisabledExplicitly(obj client.Object) bool {
	labels := obj.GetLabels()
	if labels != nil {
		val, exists := labels[consts.OdigosInstrumentationLabel]
		if exists && val == consts.InstrumentationDisabled {
			return true
		}
	}

	return false
}

func GetWorkloadObject(ctx context.Context, objectKey client.ObjectKey, kind WorkloadKind, kubeClient client.Client) (metav1.Object, error) {
	switch kind {
	case WorkloadKindDeployment:
		var deployment v1.Deployment
		err := kubeClient.Get(ctx, objectKey, &deployment)
		if err != nil {
			return nil, err
		}
		return &deployment, nil

	case WorkloadKindStatefulSet:
		var statefulSet v1.StatefulSet
		err := kubeClient.Get(ctx, objectKey, &statefulSet)
		if err != nil {
			return nil, err
		}
		return &statefulSet, nil

	case WorkloadKindDaemonSet:
		var daemonSet v1.DaemonSet
		err := kubeClient.Get(ctx, objectKey, &daemonSet)
		if err != nil {
			return nil, err
		}
		return &daemonSet, nil

	default:
		return nil, errors.New("failed to get workload object for kind: " + string(kind))
	}
}

func ExtractServiceNameFromAnnotations(annotations map[string]string, defaultName string) string {
	if annotations == nil {
		return defaultName
	}
	if reportedName, exists := annotations[consts.OdigosReportedNameAnnotation]; exists && reportedName != "" {
		return reportedName
	}
	return defaultName
}
