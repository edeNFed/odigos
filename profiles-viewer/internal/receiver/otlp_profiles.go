package receiver

import (
	"context"
	"fmt"
	"log"

	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
	"github.com/odigos-io/odigos/profiles-viewer/internal/storage"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/receiver/xreceiver"
)

// ProfilesConsumer consumes OTLP profiles and stores them.
type ProfilesConsumer struct {
	store *storage.ProfileStore
}

// NewProfilesConsumer creates a new profiles consumer.
func NewProfilesConsumer(store *storage.ProfileStore) *ProfilesConsumer {
	return &ProfilesConsumer{store: store}
}

// Capabilities returns the consumer capabilities.
func (c *ProfilesConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeProfiles processes incoming profile data.
func (c *ProfilesConsumer) ConsumeProfiles(ctx context.Context, pd pprofile.Profiles) error {
	for i := 0; i < pd.ResourceProfiles().Len(); i++ {
		rp := pd.ResourceProfiles().At(i)
		appID := extractAppID(rp.Resource().Attributes())

		if appID.IsEmpty() {
			log.Printf("Skipping profiles without app identity")
			continue
		}

		// Pass the original profiles object since it contains the dictionary
		if err := c.store.IngestProfiles(ctx, appID, pd); err != nil {
			log.Printf("Error ingesting profiles for %s: %v", appID.String(), err)
		}
	}
	return nil
}

// extractAppID extracts the application identity from resource attributes.
func extractAppID(attrs pcommon.Map) models.AppID {
	var appID models.AppID

	// Get namespace
	if ns, ok := attrs.Get("k8s.namespace.name"); ok {
		appID.Namespace = ns.Str()
	}

	// Check for deployment, daemonset, or statefulset in order of preference
	if name, ok := attrs.Get("k8s.deployment.name"); ok {
		appID.Kind = "deployment"
		appID.Name = name.Str()
		return appID
	}

	if name, ok := attrs.Get("k8s.daemonset.name"); ok {
		appID.Kind = "daemonset"
		appID.Name = name.Str()
		return appID
	}

	if name, ok := attrs.Get("k8s.statefulset.name"); ok {
		appID.Kind = "statefulset"
		appID.Name = name.Str()
		return appID
	}

	// Fallback to pod name if no workload found
	if name, ok := attrs.Get("k8s.pod.name"); ok {
		appID.Kind = "pod"
		appID.Name = name.Str()
	}

	return appID
}

// Run starts the OTLP profiles receiver.
func (c *ProfilesConsumer) Run(ctx context.Context, listenAddr string) error {
	// Create OTLP receiver factory
	f := otlpreceiver.NewFactory()

	cfg, ok := f.CreateDefaultConfig().(*otlpreceiver.Config)
	if !ok {
		log.Fatal("Failed to cast default config to otlpreceiver.Config")
	}

	// Configure gRPC listener
	cfg.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  listenAddr,
			Transport: confignet.TransportTypeTCP,
		},
	})

	// The OTLP receiver factory implements xreceiver.Factory for profiles support
	xf, ok := f.(xreceiver.Factory)
	if !ok {
		return fmt.Errorf("OTLP receiver factory does not support profiles")
	}

	// Create the profiles receiver
	r, err := xf.CreateProfiles(ctx, receivertest.NewNopSettings(f.Type()), cfg, c)
	if err != nil {
		return err
	}

	// Start receiver
	if err := r.Start(ctx, componenttest.NewNopHost()); err != nil {
		return err
	}

	log.Printf("OTLP profiles receiver listening on %s", listenAddr)

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Shutting down OTLP profiles receiver...")
	return r.Shutdown(context.Background())
}
