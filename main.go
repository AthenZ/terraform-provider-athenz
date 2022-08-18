package main

import (
	"flag"

	"github.com/AthenZ/terraform-provider-athenz/athenz"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: "yahoo/provider/athenz",
		ProviderFunc: athenz.Provider,
	}

	plugin.Serve(opts)
}
