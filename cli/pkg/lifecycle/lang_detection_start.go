package lifecycle

import (
	"context"

	"github.com/odigos-io/odigos/api/k8sconsts"

	"github.com/odigos-io/odigos/api/odigos/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RequestLangDetection struct {
	BaseTransition
}

var _ Transition = &RequestLangDetection{}

func (r *RequestLangDetection) From() State {
	return PreflightChecksPassed
}

func (r *RequestLangDetection) To() State {
	return LangDetectionInProgress
}

func (r *RequestLangDetection) Execute(ctx context.Context, obj client.Object, templateSpec *v1.PodTemplateSpec, isRemote bool) error {
	newSource := &v1alpha1.Source{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "source-",
		},
		Spec: v1alpha1.SourceSpec{
			Workload: k8sconsts.PodWorkload{
				Namespace: obj.GetNamespace(),
				Name:      obj.GetName(),
				Kind:      k8sconsts.WorkloadKind(obj.GetObjectKind().GroupVersionKind().Kind),
			},
		},
	}

	_, err := r.client.OdigosClient.Sources(obj.GetNamespace()).Create(ctx, newSource, metav1.CreateOptions{})
	return err
}
