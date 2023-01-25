package main

import (
	"flag"
	"github.com/keyval-dev/odigos/device-plugin/pkg"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"log"
)

func main() {
	flag.Parse()
	manager := dpm.NewManager(pkg.NewLister())
	log.Println("Starting Odigos Device Plugin ...")
	manager.Run()
}
