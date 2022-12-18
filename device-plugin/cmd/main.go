package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"time"
)

type lister struct{}

func (l *lister) GetResourceNamespace() string {
	return "odigos.io"
}

func (l *lister) Discover(pluginNameLists chan dpm.PluginNameList) {
	pluginNameLists <- []string{"instrumentation"}
}

func (l *lister) NewPlugin(s string) dpm.PluginInterface {
	return &plugin{}
}

type plugin struct{}

func (p *plugin) GetDevicePluginOptions(ctx context.Context, empty *v1beta1.Empty) (*v1beta1.DevicePluginOptions, error) {
	fmt.Println("GetDevicePluginOptions")
	return &v1beta1.DevicePluginOptions{
		PreStartRequired:                true,
		GetPreferredAllocationAvailable: false,
	}, nil
}

func (p *plugin) ListAndWatch(empty *v1beta1.Empty, server v1beta1.DevicePlugin_ListAndWatchServer) error {
	fmt.Println("ListAndWatch")
	server.Send(&v1beta1.ListAndWatchResponse{
		Devices: []*v1beta1.Device{
			{
				ID:     "example",
				Health: v1beta1.Healthy,
			},
		},
	})
	time.Sleep(1 * time.Hour)
	return nil
}

func (p *plugin) GetPreferredAllocation(ctx context.Context, request *v1beta1.PreferredAllocationRequest) (*v1beta1.PreferredAllocationResponse, error) {
	fmt.Printf("GetPreferredAllocation called with %v\n", request)
	return &v1beta1.PreferredAllocationResponse{}, nil
}

func (p *plugin) Allocate(ctx context.Context, request *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse, error) {
	fmt.Printf("Allocate called with %v\n", request)
	res := &v1beta1.AllocateResponse{}

	for range request.ContainerRequests {
		res.ContainerResponses = append(res.ContainerResponses, &v1beta1.ContainerAllocateResponse{
			Envs: map[string]string{
				"ODIGOS_INSTRUMENTATION": "true",
			},
			Mounts: []*v1beta1.Mount{
				{
					ContainerPath: "/odigos/EDEN_MOUNT",
					HostPath:      "/odigos/EDEN_MOUNT",
				},
			},
			Devices: []*v1beta1.DeviceSpec{
				{
					HostPath:      "/dev/odigos/EDEN_DEVICE",
					ContainerPath: "/dev/odigos/EDEN_DEVICE",
				},
			},
		})
	}

	return res, nil
}

func (p *plugin) PreStartContainer(ctx context.Context, request *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	fmt.Println("PreStartContainer")
	time.Sleep(1 * time.Hour)
	return &v1beta1.PreStartContainerResponse{}, nil
}

func main() {
	flag.Parse()
	dpm.NewManager(&lister{}).Run()
}
