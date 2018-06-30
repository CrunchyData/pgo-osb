package broker

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"sync"

	"github.com/crunchydata/pgo-osb/pgocmd"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"reflect"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	log.Info("NewBusinessLogic called")

	log.Infof("Options %v\n", o)

	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	return &BusinessLogic{
		async: o.Async,
		//instances:            make(map[string]*exampleInstance, 10),
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
	// Your catalog business logic goes here
	log.Info("GetCatalog called")
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "pgo-osb-service",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246c",
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
										"type": "object",
										"properties": map[string]interface{}{
											"color": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
												"enum": []string{
													"Clear",
													"Beige",
													"Grey",
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
		},
	}

	log.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision business logic goes here

	//jeff here is where you create the database instance using the passed params

	fmt.Printf("Provision called with params %v\n", request.Parameters)
	fmt.Printf("Provision called with InstanceID %d\n", request.InstanceID)
	fmt.Printf("Provision called with ServiceID %d\n", request.ServiceID)
	fmt.Printf("Provision called with PlanID %d\n", request.PlanID)
	// example implementation:
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

	fmt.Println("provision CO_USERNAME=" + request.Parameters["CO_USERNAME"].(string))
	fmt.Println("provision CO_PASSWORD=" + request.Parameters["CO_PASSWORD"].(string))
	fmt.Println("provision CO_CLUSTERNAME=" + request.Parameters["CO_CLUSTERNAME"].(string))

	fmt.Println("provision CO_APISERVER_URL=" + b.CO_APISERVER_URL)
	fmt.Println("provision CO_APISERVER_VERSION=" + b.CO_APISERVER_VERSION)

	pgocmd.CreateCluster(b.CO_APISERVER_URL, request.Parameters["CO_USERNAME"].(string), request.Parameters["CO_PASSWORD"].(string), request.Parameters["CO_CLUSTERNAME"].(string), b.CO_APISERVER_VERSION, request.InstanceID)
	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	// Your deprovision business logic goes here
	// jeff here is where you delete any bindings to this database
	// and the delete the database itself

	fmt.Printf("Deprovision called request=%v", request)
	fmt.Printf("Deprovision called broker request context=%v", c)
	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	fmt.Printf("Deprovision instanceID=%d\n", request.InstanceID)
	fmt.Printf("Deprovision request=%v\n", request)
	//instance2Delete := b.instances[request.InstanceID]
	//fmt.Infof("Deprovision instance2Delete=%v\n", instance2Delete)

	fmt.Println("Deprovision CO_APISERVER_URL=" + b.CO_APISERVER_URL)
	//fmt.Info("Deprovision CO_USERNAME=" + instance2Delete.Params["CO_USERNAME"].(string))
	//fmt.Info("Deprovision CO_PASSWORD=" + instance2Delete.Params["CO_PASSWORD"].(string))
	//fmt.Info("Deprovision CO_CLUSTERNAME=" + instance2Delete.Params["CO_CLUSTERNAME"].(string))
	fmt.Println("Deprovision CO_APISERVER_VERSION=" + b.CO_APISERVER_VERSION)

	pgocmd.DeleteCluster(b.CO_APISERVER_URL, b.CO_USERNAME, b.CO_PASSWORD, b.CO_APISERVER_VERSION, request.InstanceID)

	//delete(b.instances, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation business logic goes here
	log.Infoln("LastOperator called")
	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	// Your bind business logic goes here
	// jeff here is where you would return database credentials to an instance

	fmt.Printf("Bind called req=%v\n", request)
	fmt.Printf("Bind called broker ctx=%v\n", c)

	// example implementation:
	//b.Lock()
	//defer b.Unlock()

	//instance, ok := b.instances[request.InstanceID]
	//if !ok {
	//return nil, osb.HTTPStatusCodeError{
	//StatusCode: http.StatusNotFound,
	//}
	//}

	credentials, err := pgocmd.GetClusterCredentials(b.CO_APISERVER_URL, b.CO_USERNAME, b.CO_PASSWORD, b.CO_APISERVER_VERSION, request.InstanceID)
	if err != nil {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: credentials,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	fmt.Println("Unbind called")
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	fmt.Println("Update called")
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}

// example types

// exampleInstance is intended as an example of a type that holds information about a service instance
type exampleInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

func (i *exampleInstance) Match(other *exampleInstance) bool {
	return reflect.DeepEqual(i, other)
}
