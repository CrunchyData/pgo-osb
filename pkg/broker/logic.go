package broker

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
	"github.com/crunchydata/pgo-osb/pgocmd"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	api "k8s.io/api/core/v1"
	"log"
	"net/http"
	"sync"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	log.Print("NewBusinessLogic called")

	log.Printf("Options %v\n", o)

	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	return &BusinessLogic{
		async:                o.Async,
		PGO_OSB_GUID:         o.PGO_OSB_GUID,
		CO_APISERVER_URL:     o.CO_APISERVER_URL,
		CO_APISERVER_VERSION: o.CO_APISERVER_VERSION,
		CO_USERNAME:          o.CO_USERNAME,
		CO_PASSWORD:          o.CO_PASSWORD,
	}, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// Add fields here! These fields are provided purely as an example
	PGO_OSB_GUID         string
	CO_APISERVER_URL     string
	CO_APISERVER_VERSION string
	CO_USERNAME          string
	CO_PASSWORD          string
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {

	log.Print("GetCatalog called")
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name: "pgo-osb-service",
				//ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246c",
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
											"CO_USERNAME": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
											},
											"CO_CLUSTERNAME": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
											},
											"CO_PASSWORD": map[string]interface{}{
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

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {

	log.Printf("Provision called with params %v\n", request.Parameters)
	log.Printf("Provision called with InstanceID %d\n", request.InstanceID)
	log.Printf("Provision called with ServiceID %d\n", request.ServiceID)
	log.Printf("Provision called with PlanID %d\n", request.PlanID)

	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	/**
	exampleInstance := &exampleInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(exampleInstance) {
			response.Exists = true
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		}
	}
	b.instances[request.InstanceID] = exampleInstance
	*/

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	log.Print("provision CO_USERNAME=" + request.Parameters["CO_USERNAME"].(string))
	log.Print("provision CO_PASSWORD=" + request.Parameters["CO_PASSWORD"].(string))
	log.Print("provision CO_CLUSTERNAME=" + request.Parameters["CO_CLUSTERNAME"].(string))

	log.Print("provision CO_APISERVER_URL=" + b.CO_APISERVER_URL)
	log.Print("provision CO_APISERVER_VERSION=" + b.CO_APISERVER_VERSION)

	pgocmd.CreateCluster(b.CO_APISERVER_URL, request.Parameters["CO_USERNAME"].(string), request.Parameters["CO_PASSWORD"].(string), request.Parameters["CO_CLUSTERNAME"].(string), b.CO_APISERVER_VERSION, request.InstanceID)
	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {

	log.Printf("Deprovision called request=%v", request)
	log.Printf("Deprovision called broker request context=%v", c)

	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	log.Printf("Deprovision instanceID=%d\n", request.InstanceID)
	log.Printf("Deprovision request=%v\n", request)
	log.Printf("Deprovision CO_APISERVER_URL=" + b.CO_APISERVER_URL)
	log.Printf("Deprovision CO_APISERVER_VERSION=" + b.CO_APISERVER_VERSION)

	pgocmd.DeleteCluster(b.CO_APISERVER_URL, b.CO_USERNAME, b.CO_PASSWORD, b.CO_APISERVER_VERSION, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	log.Println("LastOperator called")
	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {

	log.Printf("Bind called req=%v\n", request)
	log.Printf("Bind called broker ctx=%v\n", c)

	//b.Lock()
	//defer b.Unlock()

	//instance, ok := b.instances[request.InstanceID]
	//if !ok {
	//return nil, osb.HTTPStatusCodeError{
	//StatusCode: http.StatusNotFound,
	//}
	//}

	credentials, services, err := pgocmd.GetClusterCredentials(b.CO_APISERVER_URL, b.CO_USERNAME, b.CO_PASSWORD, b.CO_APISERVER_VERSION, request.InstanceID)
	if err != nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	/**
	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: credentials,
		},
	}
	*/

	//code from kibosh example
	secretsMap := []map[string]interface{}{}
	/**
	for _, secret := range secrets.Items {
		if secret.Type == api_v1.SecretTypeOpaque {
			credentialSecrets := map[string]string{}
			for key, val := range secret.Data {
				credentialSecrets[key] = string(val)
			}
			credential := map[string]interface{}{
				"name": secret.Name,
				"data": credentialSecrets,
			}
			secretsMap = append(secretsMap, credential)
		}
	}
	*/
	//a hacked up example to see if this works with pcf
	credential := map[string]interface{}{
		"name": "somesecretname",
		"data": credentials,
	}
	secretsMap = append(secretsMap, credential)

	servicesMap := []map[string]interface{}{}
	for _, service := range services {
		spec := api.ServiceSpec{}
		spec.Ports = make([]api.ServicePort, 1)
		spec.Ports[0].Name = "postgres"
		spec.Ports[0].Port = 5432
		spec.ClusterIP = service.ClusterIP
		spec.LoadBalancerIP = service.ExternalIP
		spec.ExternalIPs = make([]string, 1)
		spec.ExternalIPs[0] = service.ExternalIP

		credentialService := map[string]interface{}{
			"name":   service.Name,
			"spec":   spec, //need to return this from the pgo call
			"status": "",   //need to return this from the pgo call
		}
		servicesMap = append(servicesMap, credentialService)
	}
	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{
				"secrets":  secretsMap,
				"services": servicesMap,
			},
		},
	}
	//end of code from kibosh example

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {

	log.Print("Unbind called")
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {

	log.Print("Update called")
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}
