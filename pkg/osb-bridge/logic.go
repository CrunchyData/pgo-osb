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
	"net/url"
	"os"
	"sync"

	"github.com/crunchydata/pgo-osb/pkg/broker"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	osblib "github.com/pmorie/osb-broker-lib/pkg/broker"
	"k8s.io/client-go/rest"
)

// Verify BusinessLogic implements the interface
var _ osblib.Interface = &BusinessLogic{}

// BusinessLogic provides an implementation of the osblib.BusinessLogic
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
	Broker                broker.Executor
	kubeAPIClient         *rest.RESTClient
}

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
		kubeAPIClient:         o.KubeAPIClient,
	}

	if o.Simulated {
		logic.Broker = broker.NewMock()
	} else {
		log.Println("Establishing remote...")
		log.Println("  PGO_APISERVER_URL=" + logic.PGO_APISERVER_URL)
		log.Println("  PGO_APISERVER_VERSION=" + logic.PGO_APISERVER_VERSION)
		log.Println("  PGO_USERNAME=" + logic.PGO_USERNAME)
		log.Println("  PGO_PASSWORD=" + logic.PGO_PASSWORD)

		r, err := broker.NewPGOperator(
			logic.kubeAPIClient,
			logic.PGO_APISERVER_URL,
			logic.PGO_USERNAME,
			logic.PGO_PASSWORD,
			logic.PGO_APISERVER_VERSION)
		if err != nil {
			log.Printf("error establishing PGO broker: %s", err)
			return nil, err
		}
		logic.Broker = r
	}

	return logic, nil
}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *osblib.RequestContext) (*osblib.CatalogResponse, error) {
	log.Println("GetCatalog called")
	response := &osblib.CatalogResponse{}
	paramSchemas := &osb.Schemas{
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
	}

	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:        "pgo-osb-service",
				ID:          b.PGO_OSB_GUID,
				Description: "The pgo osb!",
				Bindable:    true,
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
						Schemas:     paramSchemas,
					},
					{
						Name:        "standalone_sm",
						ID:          "885a1cb6-ca42-43e9-a725-8195918e1343",
						Description: "Small postgres server, no replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
					},
					{
						Name:        "standalone_md",
						ID:          "dc951396-bb28-45a4-b040-cfe3bebc6121",
						Description: "Medium postgres server, no replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
					},
					{
						Name:        "standalone_lg",
						ID:          "04349656-4dc9-4b67-9b15-52a93d64d566",
						Description: "Large postgres server, no replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
					},
					{
						Name:        "ha_sm",
						ID:          "877432f8-07eb-4e57-b984-d025a71d2282",
						Description: "Small postgres server with replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
					},
					{
						Name:        "ha_md",
						ID:          "89bcdf8a-e637-4bb3-b7ce-aca083cc1e69",
						Description: "Medium postgres server with replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
					},
					{
						Name:        "ha_lg",
						ID:          "470ca1a0-2763-41f1-a4cf-985acdb549ab",
						Description: "Large postgres server with replicas",
						Free:        truePtr(),
						Schemas:     paramSchemas,
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
	if rp.ClusterName == "" {
		return nil, fmt.Errorf("Missing required parameter: PGO_CLUSTERNAME")
	}
	if rp.Namespace == "" {
		return nil, fmt.Errorf("Missing required parameter: PGO_NAMESPACE")
	}

	log.Println("provision PGO_CLUSTERNAME=" + rp.ClusterName)
	log.Println("provision PGO_NAMESPACE=" + rp.Namespace)

	err := b.Broker.CreateCluster(broker.CreateRequest{
		InstanceID: request.InstanceID,
		Name:       rp.ClusterName,
		Namespace:  rp.Namespace,
		PlanID:     request.PlanID,
	})
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

	response := &osblib.DeprovisionResponse{}

	log.Printf("Deprovision instanceID=%s\n", request.InstanceID)
	err := b.Broker.DeleteCluster(request.InstanceID)
	if err != nil {
		if _, ok := err.(broker.ErrNoInstance); ok {
			log.Printf("Cannot find instance %s: suppressing error until HTTP 410 (Gone) can be provided", request.InstanceID)
			return response, nil
		} else {
			log.Printf("error deleting cluster: %s\n", err)
			return nil, err
		}
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *osblib.RequestContext) (*osblib.LastOperationResponse, error) {
	log.Println("LastOperation called")
	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *osblib.RequestContext) (*osblib.BindResponse, error) {
	log.Printf("Bind called req=%#v\n", request)
	log.Printf("Bind called request instanceID=%s\n", request.InstanceID)
	log.Printf("Bind called broker ctx=%#v\n", c)

	clusterDetail, err := b.Broker.ClusterDetail(request.InstanceID)
	if err != nil {
		log.Printf("error getting cluster info: %s\n", err)
		return nil, err
	}

	appID := ""
	if request.AppGUID != nil {
		appID = *request.AppGUID
	}
	bindCreds, err := b.Broker.CreateBinding(request.InstanceID, request.BindingID, appID)
	if err != nil {
		log.Printf("error getting binding info: %s\n", err)
		return nil, err
	}

	if os.Getenv("CRUNCHY_DEBUG") == "true" {
		log.Printf("credentials: %#v\n", bindCreds)
	}

	port := 5432
	dbName := clusterDetail.Database
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
				"uri": (&url.URL{
					Scheme: "postgresql",
					Host:   fmt.Sprintf("%s:%d", host, port),
					User:   url.UserPassword(bindCreds.Username, bindCreds.Password),
					Path:   dbName,
				}).String(),
			},
		},
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	if os.Getenv("CRUNCHY_DEBUG") == "true" {
		log.Printf("Bind Response: %#v\n", response)
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *osblib.RequestContext) (*osblib.UnbindResponse, error) {
	log.Printf("Unbind called req=%#v\n", request)
	err := b.Broker.DeleteBinding(request.InstanceID, request.BindingID)

	if err != nil {
		log.Printf("error during unbind: %s\n", err)
	}

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
