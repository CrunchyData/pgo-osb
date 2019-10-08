package bridge

/*
Copyright 2018 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"flag"

	"k8s.io/client-go/rest"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath           string
	PGO_OSB_GUID          string
	PGO_USERNAME          string
	PGO_PASSWORD          string
	PGO_APISERVER_URL     string
	PGO_APISERVER_VERSION string
	Async                 bool

	// Unflagged configs
	Simulated     bool
	KubeAPIClient *rest.RESTClient
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog")
	flag.StringVar(&o.PGO_APISERVER_URL, "PGO_APISERVER_URL", "", "The url to the pgo apiserver")
	flag.StringVar(&o.PGO_APISERVER_VERSION, "PGO_APISERVER_VERSION", "", "The version of the pgo apiserver")
	flag.StringVar(&o.PGO_USERNAME, "PGO_USERNAME", "", "The pgo basic auth username to authenticate with ")
	flag.StringVar(&o.PGO_PASSWORD, "PGO_PASSWORD", "", "The pgo basic auth password to authenticate with ")
	flag.StringVar(&o.PGO_OSB_GUID, "PGO_OSB_GUID", "", "The service broker guid to use for this broker instance")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")

}
