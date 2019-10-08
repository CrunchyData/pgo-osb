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
	"log"
)

// Many of the types present represent safe access to request payloads and
// provide hooks for centralized unpacking and validation
type provReqParams struct {
	ClusterName string
	Namespace   string
}

// NewProvReqParams encapsulates the parameter processing for incoming
// provision requests
func NewProvReqParams(params map[string]interface{}) *provReqParams {
	rp := &provReqParams{}
	if params == nil {
		return rp
	}

	if cn, ok := params["PGO_CLUSTERNAME"]; ok {
		if _, ok := cn.(string); ok {
			rp.ClusterName = cn.(string)
		} else {
			log.Printf("Expected type string for PGO_CLUSTERNAME, got %T instead", cn)
		}
	}

	if ns, ok := params["PGO_NAMESPACE"]; ok {
		if _, ok := ns.(string); ok {
			rp.Namespace = ns.(string)
		} else {
			log.Printf("Expected type string for PGO_NAMESPACE, got %T instead", ns)
		}
	}

	return rp
}
