package lifecycle

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/odigos-io/odigos/api/odigos/v1alpha1"

	"github.com/odigos-io/odigos/k8sutils/pkg/describe/source"

	"github.com/odigos-io/odigos/cli/pkg/remote"

	k8sutils "github.com/odigos-io/odigos/k8sutils/pkg/utils"

	"github.com/odigos-io/odigos/common"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/odigos-io/odigos/k8sutils/pkg/workload"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/odigos-io/odigos/cli/cmd/resources"

	"github.com/odigos-io/odigos/cli/pkg/kube"
	v1 "k8s.io/api/core/v1"
)

type State string

const (
	UnknownState              State = "Unknown"
	NotInstrumentedState      State = "NotInstrumented"
	PreflightChecksPassed     State = "PreflightChecksPassed"
	LangDetectionInProgress   State = "LangDetectionInProgress"
	InstrumentationInProgress State = "InstrumentationInProgress"
	InstrumentedState         State = "Instrumented"
)

type PodTemplateSpecFetcher func(ctx context.Context, name string, namespace string) (*v1.PodTemplateSpec, error)

type Orchestrator struct {
	Client          *kube.Client
	OdigosNamespace string
	TransitionsMap  map[string]Transition
	Remote          bool
}

func NewOrchestrator(client *kube.Client, ctx context.Context, isRemote bool) (*Orchestrator, error) {
	ns, err := resources.GetOdigosNamespace(client, ctx)
	if err != nil {
		return nil, err
	}

	transitions := make(map[string]Transition)
	for _, t := range allTransitions {
		t.Init(client)
		transitions[string(t.From())] = t
	}

	return &Orchestrator{Client: client,
		OdigosNamespace: ns,
		TransitionsMap:  transitions,
		Remote:          isRemote,
	}, nil
}

func (o *Orchestrator) Apply(ctx context.Context, obj client.Object, templateSpecFetcher PodTemplateSpecFetcher) error {
	// Create a channel to handle cancellation
	done := make(chan struct{})
	var finalErr error

	go func() {
		defer close(done)
		state := o.getCurrentState(ctx, obj)
		logger := slog.With("name", obj.GetName(), "namespace", obj.GetNamespace())
		logger.Info("Calculated current state", "state", state)

		if state == UnknownState {
			if err := o.rollBack(obj); err != nil {
				logger.Error("Error rolling back from unknown state", "error", err)
				finalErr = fmt.Errorf("failed to rollback from unknown state: %w", err)
				return
			}
			return
		}

		nextTransition := o.TransitionsMap[string(state)]
		for nextTransition != nil {
			select {
			case <-ctx.Done():
				// Context was cancelled, perform rollback
				logger.Warn("Context cancelled, rolling back current object")
				if err := o.rollBack(obj); err != nil {
					logger.Error("Error rolling back after context cancellation", "error", err)
					finalErr = fmt.Errorf("failed to rollback after context cancellation: %w", err)
					return
				}
				finalErr = ctx.Err()
				return
			default:
				templateSpec, err := templateSpecFetcher(ctx, obj.GetName(), obj.GetNamespace())
				if err != nil {
					slog.Error("Error fetching pod template spec", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
					finalErr = fmt.Errorf("failed to fetch template spec during transition: %w", err)
					return
				}

				if err := nextTransition.Execute(ctx, obj, templateSpec, o.Remote); err != nil {
					slog.Error("Error executing transition", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
					// Attempt rollback on execution error
					if rbErr := o.rollBack(obj); rbErr != nil {
						slog.Error("Error rolling back after failed execution", "error", rbErr, "name", obj.GetName(), "namespace", obj.GetNamespace())
						finalErr = fmt.Errorf("failed to rollback after execution error: %w", rbErr)
						return
					}
					finalErr = fmt.Errorf("failed to execute transition: %w", err)
					return
				}

				// Special case: PreflightCheck change state manually
				if nextTransition.To() == PreflightChecksPassed {
					state = PreflightChecksPassed
				} else {
					state = o.getCurrentState(ctx, obj)
				}

				slog.Info("Calculated next state", "name", obj.GetName(), "namespace", obj.GetNamespace(), "state", state)
				if state == UnknownState {
					if err := o.rollBack(obj); err != nil {
						slog.Error("Error rolling back from unknown state during transition", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
						finalErr = fmt.Errorf("failed to rollback from unknown state during transition: %w", err)
						return
					}
					return
				}

				nextTransition = o.TransitionsMap[string(state)]
			}
		}
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		// Wait for the goroutine to finish rollback
		<-done
		if finalErr == nil {
			return ctx.Err()
		}
		return finalErr
	case <-done:
		return finalErr
	}
}

func (o *Orchestrator) getCurrentState(ctx context.Context, obj client.Object) State {
	name := obj.GetName()
	kind := workload.WorkloadKindFromClientObject(obj)
	icName := workload.CalculateWorkloadRuntimeObjectName(name, kind)
	var describe *source.SourceAnalyze
	var err error
	if o.Remote {
		describe, err = remote.DescribeSource(ctx, o.Client, o.OdigosNamespace, string(kind), obj.GetNamespace(), name)
		if err != nil {
			slog.Error("Error describing source", "error", err, "name", name, "namespace", obj.GetNamespace())
			return UnknownState
		}

		if describe == nil {
			slog.Info("Describe source returned nil, skipping", "name", name, "namespace", obj.GetNamespace())
			return UnknownState
		}

		if (describe.SourceObjectsAnalysis.Workload == nil || describe.SourceObjectsAnalysis.Workload.Value == "unset") &&
			(describe.SourceObjectsAnalysis.Namespace == nil || describe.SourceObjectsAnalysis.Namespace.Value == "unset") {
			return NotInstrumentedState
		}
	}

	if !o.Remote {
		sources, err := o.Client.OdigosClient.Sources(obj.GetNamespace()).List(ctx, metav1.ListOptions{
			LabelSelector: v1alpha1.GetSourceLabelSelector(obj).String(),
		})
		if err != nil {
			slog.Error("Error listing sources", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
			return UnknownState
		}

		if sources == nil || len(sources.Items) == 0 {
			return NotInstrumentedState
		}

		_, err = o.Client.OdigosClient.InstrumentationConfigs(obj.GetNamespace()).Get(ctx, icName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return LangDetectionInProgress
			}

			slog.Error("Error getting instrumentation config", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
			return UnknownState
		}
	} else {
		if describe.RuntimeInfo == nil || len(describe.RuntimeInfo.Containers) == 0 {
			return LangDetectionInProgress
		}
	}

	if !o.Remote {
		ic, err := o.Client.OdigosClient.InstrumentationConfigs(obj.GetNamespace()).Get(ctx, icName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return LangDetectionInProgress
			}

			slog.Error("Error getting instrumented application", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
			return UnknownState
		}

		if ic.Status.RuntimeDetailsByContainer == nil || len(ic.Status.RuntimeDetailsByContainer) == 0 {
			return LangDetectionInProgress
		}

		langFound := false
		for _, rd := range ic.Status.RuntimeDetailsByContainer {
			if rd.Language != common.UnknownProgrammingLanguage && rd.Language != common.IgnoredProgrammingLanguage {
				langFound = true
				break
			}
		}

		if !langFound {
			slog.Error("Failed to detect language", "name", icName, "namespace", obj.GetNamespace())
			return UnknownState
		}
	} else {
		if describe.RuntimeInfo == nil {
			return LangDetectionInProgress
		}

		if len(describe.RuntimeInfo.Containers) == 0 {
			return LangDetectionInProgress
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
			slog.Error("Failed to detect language", "name", name, "namespace", obj.GetNamespace())
			return UnknownState
		}
	}

	instrumented, err := k8sutils.VerifyAllPodsAreInstrumented(ctx, o.Client, obj)
	if err != nil {
		slog.Error("Error verifying all pods are instrumented", "error", err, "name", obj.GetName(), "namespace", obj.GetNamespace())
		return UnknownState
	}

	if !instrumented {
		return InstrumentationInProgress
	}

	// TODO(edenfed): If relevant language + InstrumentationInstance does not exists = InstrumentationInProgress

	return InstrumentedState
}
