package pkg

import (
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"log"
)

type lister struct{}

func (l *lister) GetResourceNamespace() string {
	return "odigos.io"
}

func (l *lister) Discover(pluginNameLists chan dpm.PluginNameList) {
	pluginNameLists <- []string{"instrumentation"}
}

func (l *lister) NewPlugin(s string) dpm.PluginInterface {
	log.Printf("NewPlugin: %s", s)
	return NewInstrumentationPlugin()
}

func NewLister() dpm.ListerInterface {
	return &lister{}
}
