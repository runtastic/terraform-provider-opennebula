package main

import (
	"github.com/akurz/terraform-provider-opennebula/opennebula"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: opennebula.Provider,
	})
}
