package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/runtastic/terraform-provider-opennebula/opennebula"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: opennebula.Provider,
	})
}
