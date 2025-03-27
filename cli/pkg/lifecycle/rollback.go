package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/odigos-io/odigos/cli/pkg/remote"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/odigos-io/odigos/api/odigos/v1alpha1"
	"github.com/odigos-io/odigos/cli/pkg/kube"

	"k8s.io/apimachinery/pkg/types"

	"github.com/odigos-io/odigos/common/consts"

	"github.com/odigos-io/odigos/k8sutils/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (o *Orchestrator) rollBack(obj client.Object) error {
	// We create a new context for the rollback operation to ensure that the operation is not cancelled by the parent context
	ctx := context.Background()
	logger := slog.With("name", obj.GetName(), "namespace", obj.GetNamespace())

	logger.Info("Rolling back changes")
	if !o.Remote {
		source, err := getSource(ctx, o.Client, obj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Warn("No changes made by Odigos, skipping rollback")
				return nil
			}
			return err
		}

		err = o.Client.OdigosClient.Sources(obj.GetNamespace()).Delete(ctx, source.GetName(), metav1.DeleteOptions{})
		if err != nil {
			logger.Error("Error deleting source", "error", err)
			return err
		}
	} else {
		err := remote.DeleteSource(ctx, obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
		if err != nil {
			logger.Error("Error deleting source", "error", err)
			return err
		}
	}

	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 30*time.Minute, true, func(ctx context.Context) (bool, error) {
		rolloutCompleted, err := utils.VerifyAllPodsAreNOTInstrumented(ctx, o.Client, obj)
		if err != nil {
			logger.Error("Error verifying all pods are not instrumented", "error", err)
			return false, err
		}

		if rolloutCompleted {
			logger.Info("Rollout completed, all running pods does not contains instrumentation")
		}

		return rolloutCompleted, nil
	})

	if err != nil {
		logger.Error("Error verifying all pods are not instrumented", "error", err)
		return err
	}

	logger.Info("Rollback completed successfully")
	return nil
}

func patchOdigosLabel(ctx context.Context, client kubernetes.Interface, obj client.Object) error {
	labels := obj.GetLabels()
	if labels != nil {
		if _, ok := labels[consts.OdigosInstrumentationLabel]; !ok {
			return nil
		}
	}
	patch := fmt.Sprintf(`{"metadata":{"labels":{"%s":null}}}`, consts.OdigosInstrumentationLabel)

	switch obj.(type) {
	case *appsv1.Deployment:
		_, err := client.AppsV1().Deployments(obj.GetNamespace()).Patch(
			ctx,
			obj.GetName(),
			types.MergePatchType,
			[]byte(patch),
			metav1.PatchOptions{},
		)
		if err != nil {
			return err
		}
	case *appsv1.StatefulSet:
		_, err := client.AppsV1().StatefulSets(obj.GetNamespace()).Patch(
			ctx,
			obj.GetName(),
			types.MergePatchType,
			[]byte(patch),
			metav1.PatchOptions{},
		)
		if err != nil {
			return err
		}
	case *appsv1.DaemonSet:
		_, err := client.AppsV1().DaemonSets(obj.GetNamespace()).Patch(
			ctx,
			obj.GetName(),
			types.MergePatchType,
			[]byte(patch),
			metav1.PatchOptions{},
		)
		if err != nil {
			return err
		}
	}
	return nil

}

func getSource(ctx context.Context, c *kube.Client, obj client.Object) (*v1alpha1.Source, error) {
	sources, err := c.OdigosClient.Sources(obj.GetNamespace()).List(ctx, metav1.ListOptions{
		LabelSelector: v1alpha1.GetSourceLabelSelector(obj).String(),
	})

	if err != nil {
		return nil, err
	}

	if sources == nil || len(sources.Items) != 1 {
		return nil, fmt.Errorf("expected 1 source, got %d", len(sources.Items))
	}

	return &sources.Items[0], nil
}
