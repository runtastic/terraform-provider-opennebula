package main

import (
	"github.com/hashicorp/terraform/plugin"
	"stash.runtastic.com/ropscode/opennebula-provider/opennebula"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: opennebula.Provider,
	})
}
