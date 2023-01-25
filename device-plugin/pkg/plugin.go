package pkg

import (
	"context"
	"fmt"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"log"
)

type plugin struct {
	stopCh chan struct{}
}

func NewInstrumentationPlugin() dpm.PluginInterface {
	return &plugin{
		stopCh: make(chan struct{}),
	}
}

func (p *plugin) GetDevicePluginOptions(ctx context.Context, empty *v1beta1.Empty) (*v1beta1.DevicePluginOptions, error) {
	fmt.Println("GetDevicePluginOptions")
	return &v1beta1.DevicePluginOptions{
		PreStartRequired:                true,
		GetPreferredAllocationAvailable: true,
	}, nil
}

func (p *plugin) ListAndWatch(empty *v1beta1.Empty, server v1beta1.DevicePlugin_ListAndWatchServer) error {
	server.Send(&v1beta1.ListAndWatchResponse{
		Devices: []*v1beta1.Device{
			{
				ID:     "instrumentation-device",
				Health: v1beta1.Healthy,
			},
			{
				ID:     "instrumentation-device2",
				Health: v1beta1.Healthy,
			},
		},
	})

	<-p.stopCh
	return nil
}

func (p *plugin) Stop() error {
	log.Println("Stopping Odigos Device Plugin ...")
	return nil
}

func (p *plugin) GetPreferredAllocation(ctx context.Context, request *v1beta1.PreferredAllocationRequest) (*v1beta1.PreferredAllocationResponse, error) {
	log.Printf("GetPreferredAllocation request: %v", request)
	return &v1beta1.PreferredAllocationResponse{}, nil
}

func (p *plugin) Allocate(ctx context.Context, request *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse, error) {
	log.Printf("Allocate request: %v", request)

	err := getPodResources()
	if err != nil {
		return nil, err
	}

	res := &v1beta1.AllocateResponse{}

	for range request.ContainerRequests {
		res.ContainerResponses = append(res.ContainerResponses, &v1beta1.ContainerAllocateResponse{
			Envs: map[string]string{
				"ODIGOS_INSTRUMENTATION": "true",
				"EDEN_TEST":              "${HOSTNAME}",
			},
			Mounts: []*v1beta1.Mount{
				{
					ContainerPath: "/odigos/EDEN_MOUNT",
					HostPath:      "/odigos/EDEN_MOUNT",
				},
			},
		})
	}

	return res, nil
}

func (p *plugin) PreStartContainer(ctx context.Context, request *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	log.Printf("PreStartContainer request: %v", request)
	return &v1beta1.PreStartContainerResponse{}, nil
}
