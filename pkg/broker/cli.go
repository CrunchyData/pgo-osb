package broker

import (
	"flag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath          string
	PGO_OSB_GUID         string
	CO_USERNAME          string
	CO_PASSWORD          string
	CO_APISERVER_URL     string
	CO_APISERVER_VERSION string
	Async                bool
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog")
	flag.StringVar(&o.CO_APISERVER_URL, "CO_APISERVER_URL", "", "The url to the pgo apiserver")
	flag.StringVar(&o.CO_APISERVER_VERSION, "CO_APISERVER_VERSION", "", "The version of the pgo apiserver")
	flag.StringVar(&o.CO_USERNAME, "CO_USERNAME", "", "The pgo basic auth username to authenticate with ")
	flag.StringVar(&o.CO_PASSWORD, "CO_PASSWORD", "", "The pgo basic auth password to authenticate with ")
	flag.StringVar(&o.PGO_OSB_GUID, "PGO_OSB_GUID", "", "The service broker guid to use for this broker instance")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")

}
