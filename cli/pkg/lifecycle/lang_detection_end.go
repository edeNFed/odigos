package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/odigos-io/odigos/cli/cmd/resources"

	"github.com/odigos-io/odigos/cli/pkg/remote"

	"github.com/odigos-io/odigos/common"
	"github.com/odigos-io/odigos/k8sutils/pkg/workload"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WaitForLangDetection struct {
	BaseTransition
}

func (w *WaitForLangDetection) From() State {
	return LangDetectionInProgress
}

func (w *WaitForLangDetection) To() State {
	return InstrumentationInProgress
}

func (w *WaitForLangDetection) Execute(ctx context.Context, obj client.Object, templateSpec *corev1.PodTemplateSpec, isRemote bool) error {
	name := obj.GetName()
	kind := workload.WorkloadKindFromClientObject(obj)
	iaName := workload.CalculateWorkloadRuntimeObjectName(name, kind)
	return wait.PollUntilContextTimeout(ctx, 1*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		if !isRemote {
			ic, err := w.client.OdigosClient.InstrumentationConfigs(obj.GetNamespace()).Get(ctx, iaName, metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					w.log("Error while fetching InstrumentationConfig: " + err.Error())
				}
				return false, nil
			}

			if ic.Status.RuntimeDetailsByContainer == nil || len(ic.Status.RuntimeDetailsByContainer) == 0 {
				return false, nil
			}

			langFound := false
			for _, rd := range ic.Status.RuntimeDetailsByContainer {
				if rd.Language != common.UnknownProgrammingLanguage && rd.Language != common.IgnoredProgrammingLanguage {
					w.log(fmt.Sprintf("Detected language: %s", rd.Language))
					langFound = true
					break
				}
			}

			if !langFound {
				return false, errors.New("Failed to detect language")
			}

			return true, nil
		}

		ns, err := resources.GetOdigosNamespace(w.client, ctx)
		if err != nil {
			return false, nil
		}

		describe, err := remote.DescribeSource(ctx, w.client, ns, string(kind), obj.GetNamespace(), name)
		if describe == nil {
			return false, nil
		}

		if err != nil {
			w.log(fmt.Sprintf("Error describing source: %s", err))
			return false, nil
		}

		if describe.RuntimeInfo == nil {
			return false, nil
		}

		if len(describe.RuntimeInfo.Containers) == 0 {
			return false, nil
		}

		langFound := false
		for _, c := range describe.RuntimeInfo.Containers {
			langStr, ok := c.Language.Value.(string)
			if !ok {
				continue
			}

			langParsed := common.ProgrammingLanguage(langStr)
			if langParsed != common.UnknownProgrammingLanguage && langParsed != common.IgnoredProgrammingLanguage {
				langFound = true
				break
			}
		}

		if !langFound {
			return false, errors.New("Failed to detect language")
		}

		return true, nil
	})
}

var _ Transition = &WaitForLangDetection{}
