/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/log"

	"golang.org/x/sync/errgroup"

	"github.com/odigos-io/odigos/cli/cmd/resources"

	"github.com/odigos-io/odigos/api/k8sconsts"

	"github.com/odigos-io/odigos/cli/pkg/remote"

	corev1 "k8s.io/api/core/v1"

	"github.com/odigos-io/odigos/cli/pkg/lifecycle"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/odigos-io/odigos/cli/pkg/kube"

	"github.com/odigos-io/odigos/cli/pkg/preflight"

	"github.com/spf13/cobra"
)

const (
	excludeNamespacesFileFlag  = "exclude-namespaces-file"
	excludeAppsFileFlag        = "exclude-apps-file"
	skipPreflightCheckFlag     = "skip-preflight-checks"
	dryRunFlag                 = "dry-run"
	instrumentationCollOffFlag = "instrumentation-cool-off"
	remoteFlag                 = "remote"
	onlyDeploymentFlag         = "only-deployment"
	onlyNamespaceFlag          = "only-namespace"
	goroutinesFlag             = "goroutines"
)

// instrumentCmd represents the instrument command
var instrumentCmd = &cobra.Command{
	Use:   "instrument",
	Short: "Instrument applications with Odigos",
	Long: `Instrument applications with Odigos. This command will instrument the application with Odigos CLI
and monitor the instrumentation status.`,
}

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Instrument entire cluster with Odigos",
	Long: `Instrument entire cluster with Odigos. This command will instrument the entire Kubernetes cluster with
Odigos CLI and monitor the instrumentation status.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.SetDefault(slog.New(log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: true,
			TimeFormat:      time.Kitchen,
		})))

		ctx, cancel := context.WithCancel(cmd.Context())
		var uiClient *remote.UIClientViaPortForward
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		defer func() {
			if uiClient != nil {
				uiClient.Close()
			}
		}()
		defer signal.Stop(ch)

		go func() {
			<-ch
			cancel()
		}()

		parallelSizeStr := cmd.Flag(goroutinesFlag).Value.String()
		parallelSize, err := strconv.Atoi(parallelSizeStr)
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Invalid value for goroutines: %s\n", err))
			return
		}

		group, ctx := errgroup.WithContext(ctx)
		group.SetLimit(parallelSize)

		excludedNs, err := readFileLines(cmd.Flag(excludeNamespacesFileFlag).Value.String())
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot read exclude-namespaces-file: %s\n", err))
			return
		}

		excludedApps, err := readFileLines(cmd.Flag(excludeAppsFileFlag).Value.String())
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot read exclude-apps-file: %s\n", err))
			return
		}

		dryRun := cmd.Flag(dryRunFlag).Changed && cmd.Flag(dryRunFlag).Value.String() == "true"
		isRemote := cmd.Flag(remoteFlag).Changed && cmd.Flag(remoteFlag).Value.String() == "true"
		coolOffStr := cmd.Flag(instrumentationCollOffFlag).Value.String()
		coolOff, err := time.ParseDuration(coolOffStr)
		ctx = lifecycle.SetCoolOff(ctx, coolOff)
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Invalid duration for instrumentation-cool-off: %s\n", err))
			return
		}

		onlyDeployment := cmd.Flag(onlyDeploymentFlag).Value.String()
		onlyNamespace := cmd.Flag(onlyNamespaceFlag).Value.String()

		if (onlyDeployment != "" && onlyNamespace == "") || (onlyDeployment == "" && onlyNamespace != "") {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m --only-deployment and --only-namespace must be set together\n"))
			return
		}

		fmt.Printf("About to instrument with Odigos\n")
		if dryRun {
			fmt.Printf("Dry-Run mode ENABLED - No changes will be made\n")
		}
		if onlyDeployment != "" {
			fmt.Printf("Instrumenting deployment %s in namespace %s\n", onlyDeployment, onlyNamespace)
		} else {
			fmt.Printf("Excluded Namespaces:   %d\n", len(excludedNs))
			fmt.Printf("Excluded Applications: %d\n", len(excludedApps))
		}
		fmt.Printf("%-50s", "Checking if Kubernetes cluster is reachable")
		client := kube.GetCLIClientOrExit(cmd)
		fmt.Printf("\u001B[32mPASS\u001B[0m\n\n")

		if isRemote {
			uiClient, err = remote.NewUIClient(client, ctx)
			if err != nil {
				printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot create remote UI client: %s\n", err))
				return
			}

			fmt.Println("Flag --remote is set, starting port-forward to UI pod ...")
			go func() {
				if err := uiClient.Start(); err != nil {
					printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot start remote UI client: %s\n", err))
					return
				}
			}()

			<-uiClient.Ready()
			port, err := uiClient.DiscoverLocalPort()
			if err != nil {
				printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot discover local port for UI client: %s\n", err))
				return
			}
			fmt.Printf("Remote client is using local port %s\n", port)
		}

		runPreflightChecks(ctx, cmd, client, isRemote)

		slog.Info("Starting instrumentation ...")
		instrumentCluster(ctx, client, excludedNs, excludedApps, dryRun, isRemote, onlyNamespace, onlyDeployment, group)
	},
}

func instrumentCluster(ctx context.Context, client *kube.Client, excludedNs, excludedApps map[string]struct{}, dryRun bool, remote bool, onlyNamespace, onlyDeployment string,
	group *errgroup.Group) {
	systemNs := sliceToMap(k8sconsts.DefaultIgnoredNamespaces)
	odigosNs, err := resources.GetOdigosNamespace(client, ctx)
	systemNs[odigosNs] = struct{}{}
	if err != nil {
		printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot get Odigos namespace: %s\n", err))
		return
	}

	if onlyDeployment != "" {
		orchestrator, err := lifecycle.NewOrchestrator(client, ctx, remote)
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot create orchestrator: %s\n", err))
			return
		}

		dep, err := client.AppsV1().Deployments(onlyNamespace).Get(ctx, onlyDeployment, metav1.GetOptions{})

		// TODO: This is ugly hack to make controller-runtime based functions work, need to refactor
		dep.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		}
		if err != nil {
			printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot get deployment %s in namespace %s: %s\n", onlyDeployment, onlyNamespace, err))
			return
		}

		if dryRun {
			fmt.Printf("Dry-Run mode ENABLED - No changes will be made\n")
			return
		}

		group.Go(func() error {
			err = orchestrator.Apply(ctx, dep, func(ctx context.Context, name string, namespace string) (*corev1.PodTemplateSpec, error) {
				dep, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				return &dep.Spec.Template, nil
			})

			if err != nil {
				printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Failed to instrument deployment: %s\n", err))
			}
			return nil
		})

		return
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot list namespaces: %s\n", err))
		return
	}

	orchestrator, err := lifecycle.NewOrchestrator(client, ctx, remote)
	if err != nil {
		printFatalError(err, fmt.Sprintf("\033[31mERROR\033[0m Cannot create orchestrator: %s\n", err))
		return
	}

	for _, ns := range nsList.Items {
		slog.Info("Instrumenting namespace", "namespace", ns.Name)
		_, excluded := excludedNs[ns.Name]
		_, system := systemNs[ns.Name]
		if excluded || system {
			slog.Warn("Skipping namespace due to exclusion file or system namespace", "namespace", ns.Name)
			continue
		}

		err = instrumentNamespace(ctx, client, ns.Name, excludedApps, orchestrator, dryRun, group)
		if errors.Is(err, context.Canceled) {
			return
		}
	}
}

func instrumentNamespace(ctx context.Context, client *kube.Client, ns string, excludedApps map[string]struct{}, orchestrator *lifecycle.Orchestrator, dryRun bool,
	group *errgroup.Group) error {
	deps, err := client.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("  - \033[31mERROR\033[0m Cannot list deployments: %s\n", err)
		return nil
	}

	for _, dep := range deps.Items {
		// TODO: This is ugly hack to make controller-runtime based functions work, need to refactor
		dep.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		}

		logger := slog.With("name", dep.Name, "namespace", ns)
		logger.Info("Inspecting Deployment")
		_, excluded := excludedApps[dep.Name]
		if excluded {
			logger.Warn("Skipping deployment due to exclusion file")
			continue
		}

		if dryRun {
			logger.Warn("Dry-Run mode ENABLED - No changes will be made")
			continue
		}

		group.Go(func() error {
			err = orchestrator.Apply(ctx, &dep, func(ctx context.Context, name string, namespace string) (*corev1.PodTemplateSpec, error) {
				dep, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				return &dep.Spec.Template, nil
			})

			if errors.Is(err, context.Canceled) {
				return err
			}

			return nil
		})
	}

	return nil
}

func runPreflightChecks(ctx context.Context, cmd *cobra.Command, client *kube.Client, remote bool) {
	shouldSkip := cmd.Flag(skipPreflightCheckFlag).Changed && cmd.Flag(skipPreflightCheckFlag).Value.String() == "true"
	if shouldSkip {
		fmt.Printf("Skipping preflight checks due to --%s flag\n", skipPreflightCheckFlag)
		return
	}

	fmt.Printf("Running preflight checks:\n")
	for _, check := range preflight.AllChecks {
		fmt.Printf("  - %-60s", check.Description())
		if err := check.Execute(client, ctx, remote); err != nil {
			fmt.Printf("\u001B[31mERROR\u001B[0m\n\n")
			fmt.Printf("Check failed: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("\u001B[32mPASS\u001B[0m\n")
		}
	}

	fmt.Printf("  - All preflight checks passed!\n\n")
}

func readFileLines(filePath string) (map[string]struct{}, error) {
	if filePath == "" {
		return nil, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines[scanner.Text()] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func init() {
	rootCmd.AddCommand(instrumentCmd)
	instrumentCmd.AddCommand(clusterCmd)
	clusterCmd.Flags().String(excludeNamespacesFileFlag, "", "File containing namespaces to exclude from instrumentation. Each namespace should be on a new line.")
	clusterCmd.Flags().String(excludeAppsFileFlag, "", "File containing applications to exclude from instrumentation. Each application should be on a new line.")
	clusterCmd.Flags().Bool(skipPreflightCheckFlag, false, "Skip preflight checks")
	clusterCmd.Flags().Bool(dryRunFlag, false, "Dry run mode")
	clusterCmd.Flags().Duration(instrumentationCollOffFlag, 0, "Cool-off period for instrumentation. Time format is 1h30m")
	clusterCmd.Flags().Bool(remoteFlag, false, "Use remote in-cluster service for checking instrumentation status")
	clusterCmd.Flags().String(onlyNamespaceFlag, "", "Namespace of the deployment to instrument (must be used with --only-deployment)")
	clusterCmd.Flags().String(onlyDeploymentFlag, "", "Name of the deployment to instrument (must be used with --only-namespace)")
	clusterCmd.Flags().Int(goroutinesFlag, 3, "Number of goroutines to use for parallel onboarding (default 3)")
}

func sliceToMap(slice []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range slice {
		m[s] = struct{}{}
	}
	return m
}

func printFatalError(err error, str string) {
	if errors.Is(err, context.Canceled) {
		return
	}

	fmt.Print(str)
	os.Exit(1)
}
