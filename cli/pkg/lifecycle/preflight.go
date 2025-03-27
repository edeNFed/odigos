package lifecycle

import (
	"context"
	"fmt"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PreflightCheck struct {
	BaseTransition
}

func (p *PreflightCheck) From() State {
	return NotInstrumentedState
}

func (p *PreflightCheck) To() State {
	return PreflightChecksPassed
}

func (p *PreflightCheck) Execute(ctx context.Context, obj client.Object, templateSpec *v1.PodTemplateSpec, isRemote bool) error {
	switch obj.(type) {
	case *appsv1.Deployment:
		deployment := obj.(*appsv1.Deployment)
		ru := deployment.Spec.Strategy.RollingUpdate
		if ru != nil && ru.MaxUnavailable != nil && ru.MaxUnavailable.StrVal == "100%" {
			return fmt.Errorf("Deployment %s has MaxUnavailable set to 100%%", deployment.Name)
		} else {
			slog.Info("Deployment MaxUnavailable check passed", "name", deployment.Name, "namespace", deployment.Namespace)
		}

		// Check if all pods of the deployment are healthy
		if deployment.Status.UnavailableReplicas > 0 {
			return fmt.Errorf("Deployment %s has %d unavailable replicas", deployment.Name, deployment.Status.UnavailableReplicas)
		} else {
			slog.Info("Deployment replicas check passed", "name", deployment.Name, "namespace", deployment.Namespace)
		}
	}

	return nil
}

var _ Transition = &PreflightCheck{}
