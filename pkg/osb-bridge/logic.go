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
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/crunchydata/pgo-osb/pkg/broker"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	osblib "github.com/pmorie/osb-broker-lib/pkg/broker"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	log.Println("NewBusinessLogic called")
	log.Printf("Options %#v\n", o)

	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	logic := &BusinessLogic{
		async:                 o.Async,
		PGO_OSB_GUID:          o.PGO_OSB_GUID,
		PGO_APISERVER_URL:     o.PGO_APISERVER_URL,
		PGO_APISERVER_VERSION: o.PGO_APISERVER_VERSION,
		PGO_USERNAME:          o.PGO_USERNAME,
		PGO_PASSWORD:          o.PGO_PASSWORD,
	}

	if o.Simulated {
		// NoOp for now
	} else {
		log.Println("Establishing remote...")
		log.Println("  PGO_APISERVER_URL=" + logic.PGO_APISERVER_URL)
		log.Println("  PGO_APISERVER_VERSION=" + logic.PGO_APISERVER_VERSION)
		log.Println("  PGO_USERNAME=" + logic.PGO_USERNAME)
		log.Println("  PGO_PASSWORD=" + logic.PGO_PASSWORD)

		r, err := broker.NewPGORemote(
			logic.PGO_APISERVER_URL,
			logic.PGO_USERNAME,
			logic.PGO_PASSWORD,
			logic.PGO_APISERVER_VERSION)
		if err != nil {
			log.Printf("error establishing PGORemote: %s", err)
			return nil, err
		}
		logic.Remote = r
	}

	return logic, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex

	PGO_OSB_GUID          string
	PGO_APISERVER_URL     string
	PGO_APISERVER_VERSION string
	PGO_USERNAME          string
	PGO_PASSWORD          string
	Remote                broker.Remote
}

var _ osblib.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *osblib.RequestContext) (*osblib.CatalogResponse, error) {

	log.Println("GetCatalog called")
	response := &osblib.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "pgo-osb-service",
				ID:            b.PGO_OSB_GUID,
				Description:   "The pgo osb!",
				Bindable:      true,
				PlanUpdatable: truePtr(),
				Metadata: map[string]interface{}{
					"displayName": "pgo osb service",
					"imageUrl":    "https://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "86064792-7ea2-467b-af93-ac9694d96d5c",
						Description: "The default plan for the pgo osb service",
						Free:        truePtr(),
						Schemas: &osb.Schemas{
							ServiceInstance: &osb.ServiceInstanceSchema{
								Create: &osb.InputParametersSchema{
									Parameters: map[string]interface{}{
										"type":    "object",
										"$schema": "http://json-schema.org/draft-04/schema#",
										"properties": map[string]interface{}{
											"PGO_CLUSTERNAME": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
											},
											"PGO_NAMESPACE": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	log.Printf("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *osblib.RequestContext) (*osblib.ProvisionResponse, error) {
	log.Printf("Provision called with params %#v\n", request.Parameters)
	log.Printf("Provision called with InstanceID %s\n", request.InstanceID)
	log.Printf("Provision called with ServiceID %s\n", request.ServiceID)
	log.Printf("Provision called with PlanID %s\n", request.PlanID)

	b.Lock()
	defer b.Unlock()

	response := osblib.ProvisionResponse{}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	// Since handling request.Parameters is being delegated to the
	// encapsulating type, direct access beyond here should raise suspicion
	rp := NewProvReqParams(request.Parameters)

	log.Println("provision PGO_CLUSTERNAME=" + rp.ClusterName)
	log.Println("provision PGO_NAMESPACE=" + rp.Namespace)

	err := b.Remote.CreateCluster(request.InstanceID, rp.ClusterName, rp.Namespace)
	if err != nil {
		log.Printf("error during Provision: %s", err)
		return nil, err
	}
	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *osblib.RequestContext) (*osblib.DeprovisionResponse, error) {
	log.Printf("Deprovision called request=%#v", request)
	log.Printf("Deprovision called broker request context=%#v", c)

	b.Lock()
	defer b.Unlock()

	response := osblib.DeprovisionResponse{}

	log.Printf("Deprovision instanceID=%s\n", request.InstanceID)
	err := b.Remote.DeleteCluster(request.InstanceID)
	if err != nil {
		log.Printf("error deleting cluster: %s\n", err)
		return nil, err
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *osblib.RequestContext) (*osblib.LastOperationResponse, error) {
	log.Println("LastOperation called")
	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *osblib.RequestContext) (*osblib.BindResponse, error) {
	log.Printf("Bind called req=%#v\n", request)
	log.Printf("Bind called request instanceID=%s\n", request.InstanceID)
	log.Printf("Bind called broker ctx=%#v\n", c)

	clusterDetail, err := b.Remote.ClusterDetail(request.InstanceID)
	if err != nil {
		log.Printf("error getting cluster info: %s\n", err)
		return nil, err
	}

	appGUID := ""
	if request.AppGUID != nil {
		appGUID = *request.AppGUID
	}

	bindCreds, err := b.Remote.BindingUser(request.InstanceID, appGUID, request.BindingID)
	if err != nil {
		log.Printf("error getting binding info: %s\n", err)
		return nil, err
	}

	if os.Getenv("CRUNCHY_DEBUG") == "true" {
		log.Printf("credentials: %#v\n", bindCreds)
	}

	port := 5432
	dbName := "userdb"
	host := clusterDetail.ExternalIP
	if host == "" {
		host = clusterDetail.ClusterIP
	}
	response := osblib.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{
				"username":      bindCreds.Username,
				"password":      bindCreds.Password,
				"db_port":       port,
				"db_name":       dbName,
				"db_host":       host,
				"internal_host": clusterDetail.ClusterIP,
				"uri": fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
					bindCreds.Username,
					bindCreds.Password,
					host,
					port,
					dbName),
			},
		},
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	log.Printf("Bind Response: %#v\n", response)

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *osblib.RequestContext) (*osblib.UnbindResponse, error) {
	log.Println("Unbind called")
	return &osblib.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *osblib.RequestContext) (*osblib.UpdateInstanceResponse, error) {
	log.Println("Update called")
	response := osblib.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}
