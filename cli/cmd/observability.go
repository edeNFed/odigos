package cmd

import (
	"bufio"
	"fmt"
	_ "github.com/keyval-dev/odigos/api/odigos/v1alpha1"
	odigosv1 "github.com/keyval-dev/odigos/api/odigos/v1alpha1"
	"github.com/keyval-dev/odigos/cli/cmd/observability/backend"
	"github.com/keyval-dev/odigos/cli/pkg/kube"
	"github.com/keyval-dev/odigos/common"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

// observabilityCmd represents the observability command
var observabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := isValidBackend(backendFlag); err != nil {
			return err
		}

		be := backend.Get(backendFlag)
		parsedArgs, err := be.ParseFlags(cmd)
		if err != nil {
			return err
		}

		signals, err := calculateSignals(signalsFlag, be.SupportedSignals(), be.Name())
		if err != nil {
			return err
		}

		fmt.Println("About to install the following observability skill:")
		fmt.Println("Target applications: all recognized applications")
		fmt.Println("Infra: OpenTelemetry Collector")
		fmt.Printf("Signals: %s\n", strings.Join(signalsToString(signals), ","))
		if !skipConfirm {
			confirm, err := askForConfirmation()
			if err != nil {
				return err
			}

			if !confirm {
				fmt.Println("Aborting installation.")
				return nil
			}
		}

		if err := persistArgs(parsedArgs, cmd, signals, be.Name()); err != nil {
			return err
		}

		fmt.Printf("\n\u001B[32mSUCCESS:\u001B[0m Observability skill installed.\n")
		return nil
	},
}

var (
	backendFlag string
	apiKeyFlag  string
	urlFlag     string
	signalsFlag []string
	skipConfirm bool
)

func isValidBackend(name string) error {
	avail := backend.GetAvailableBackends()
	if name == "" {
		return fmt.Errorf("please specifiy an observability backend via --backend flag, choose one from %+v", avail)
	}

	for _, s := range avail {
		if name == s {
			return nil
		}
	}

	return fmt.Errorf("invalid backend %s, choose from %+v", name, avail)
}

func persistArgs(args *backend.ObservabilityArgs, cmd *cobra.Command, signals []common.ObservabilitySignal, backendName common.DestinationType) error {
	kc := kube.CreateClient(cmd)
	ns, err := getOdigosNamespace(kc)
	if err != nil {
		return err
	}

	_, err = kc.OdigosClient.OdigosConfigurations(ns).Create(cmd.Context(), &odigosv1.OdigosConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "odigos-config",
		},
		Spec: odigosv1.OdigosConfigurationSpec{
			InstrumentationMode: odigosv1.OptOutInstrumentationMode,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	skillName := "observability" // TODO
	_, err = kc.CoreV1().Secrets(ns).Create(cmd.Context(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: skillName,
		},
		StringData: args.Secret,
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = kc.OdigosClient.Destinations(ns).Create(cmd.Context(), &odigosv1.Destination{
		ObjectMeta: metav1.ObjectMeta{
			Name: skillName,
		},
		Spec: odigosv1.DestinationSpec{
			Type: backendName,
			Data: args.Data,
			SecretRef: &v1.LocalObjectReference{
				Name: skillName,
			},
			Signals: signals,
		},
	}, metav1.CreateOptions{})

	return err
}

func calculateSignals(args []string, supported []common.ObservabilitySignal, beName common.DestinationType) ([]common.ObservabilitySignal, error) {
	if len(args) == 0 {
		return supported, nil
	}

	supportedMap := make(map[common.ObservabilitySignal]interface{}, len(supported))
	for _, s := range supported {
		supportedMap[s] = nil
	}

	var result []common.ObservabilitySignal
	for _, s := range args {
		signal, ok := common.GetSignal(s)
		if !ok {
			return nil, fmt.Errorf("%s is not a valid signal choose from %+v", s, signalsToString(supported))
		}

		if _, exists := supportedMap[signal]; !exists {
			return nil, fmt.Errorf("%s is not supported as a %s signal. Choose from the following signals %+v or choose a different backend", s, beName, signalsToString(supported))
		}

		result = append(result, signal)
	}

	return result, nil
}

func signalsToString(signals []common.ObservabilitySignal) []string {
	var result []string
	for _, s := range signals {
		result = append(result, strings.ToLower(string(s)))
	}
	return result
}

func getOdigosNamespace(kubeClient *kube.Client) (string, error) {
	// TODO: find namespace by label
	return "odigos-system", nil
}

func askForConfirmation() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to continue? [Y/n]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true, nil
	} else if response == "n" || response == "no" {
		return false, nil
	}

	return false, fmt.Errorf("%s invalid response. Type [y/n/yes/no]", response)
}

func init() {
	skillCmd.AddCommand(observabilityCmd)
	observabilityCmd.Flags().StringVar(&backendFlag, "backend", "", "Backend for observability data")
	observabilityCmd.Flags().StringVarP(&urlFlag, "url", "u", "", "URL of the backend for observability data")
	observabilityCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key for the selected backend")
	observabilityCmd.Flags().StringSliceVarP(&signalsFlag, "signal", "s", nil, "Reported signals [traces,metrics,logs]")
	observabilityCmd.Flags().BoolVarP(&skipConfirm, "no-prompt", "y", false, "Skip install confirmation")
}
