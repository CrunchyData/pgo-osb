package broker

/*
 Copyright 2017-2018 Crunchy Data Solutions, Inc.
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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	crv1 "github.com/crunchydata/postgres-operator/apis/cr/v1"
	msgs "github.com/crunchydata/postgres-operator/apiservermsgs"
	"github.com/crunchydata/postgres-operator/kubeapi"
	api "github.com/crunchydata/postgres-operator/pgo/api"

	"k8s.io/client-go/rest"
)

const (
	// Could be in New only, but reinforces the DON'T TOUCH nature
	_INSTANCE_LABEL_KEY = "pgo-osb-instance"
	_BIND_LABEL_KEY     = "pgo-osb-bindid"
)

type PGOperator struct {
	remoteURL    string
	bindLabelKey string
	clientVer    string
	instLabelKey string
	kubeClient   *rest.RESTClient
	pgoCreds     msgs.BasicAuthCredentials
	nsLookup     map[string]string
	nsMutex      sync.RWMutex
}

// NewPGOperator sets up authentication information for a PGO client
func NewPGOperator(KubeClient *rest.RESTClient, APIServerURL, basicAuthUsername, basicAuthPassword, clientVersion string) (*PGOperator, error) {
	if KubeClient == nil {
		return nil, errors.New("KubeClient cannot be nil")
	}
	po := &PGOperator{
		bindLabelKey: _BIND_LABEL_KEY,
		clientVer:    clientVersion,
		instLabelKey: _INSTANCE_LABEL_KEY,
		kubeClient:   KubeClient,
		nsLookup:     map[string]string{},
		pgoCreds: msgs.BasicAuthCredentials{
			APIServerURL: APIServerURL,
			Username:     basicAuthUsername,
			Password:     basicAuthPassword,
		},
		remoteURL: APIServerURL,
	}

	// TEST: Files there at start?
	_, err := po.httpClient()
	if err != nil {
		log.Printf("error on initial httpClient: %s", err)
		return nil, err
	}

	return po, nil
}

// findInstanceNamespace finds the cluster for a given instID to get the
// namespace for searching via the PGO API. It caches seen values to avoid
// continual kubeapi lookups
func (po *PGOperator) findInstanceNamespace(instID string) (string, error) {
	po.nsMutex.RLock()
	if ns, ok := po.nsLookup[instID]; ok {
		po.nsMutex.RUnlock()
		return ns, nil
	} else {
		po.nsMutex.RUnlock()
		selector := po.instLabel(instID)
		log.Print("find cluster " + selector)

		clusterList := crv1.PgclusterList{}
		err := kubeapi.GetpgclustersBySelector(po.kubeClient, &clusterList, selector, "")
		if err != nil {
			return "", err
		}
		if l := len(clusterList.Items); l > 1 {
			log.Printf("Found %d clusters for instance id %s, using first in list", l, instID)
		} else if l == 0 {
			log.Printf("Found no clusters for instance id %s", instID)
			return "", errors.New("cluster for instanceID " + instID + " not found by selector")
		}

		po.nsMutex.Lock()
		ns := clusterList.Items[0].GetNamespace()
		po.nsLookup[instID] = ns
		po.nsMutex.Unlock()

		return ns, nil
	}
}

// httpClient provides an http client based on the current state of bound
// apiserver-keys
// TODO: Poll cert changes and cache client between
func (po *PGOperator) httpClient() (*http.Client, error) {
	caCertPath := "/opt/apiserver-keys/ca.crt"
	clientCertPath := "/opt/apiserver-keys/client.crt"
	clientKeyPath := "/opt/apiserver-keys/client.key"

	// Set up client trust
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Printf("loading %s: %s", caCertPath, err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Initialize combined X.509 cert
	_, err = ioutil.ReadFile(clientCertPath)
	if err != nil {
		log.Printf("loading %s: %s\n", clientCertPath, err)
		return nil, err
	}
	_, err = ioutil.ReadFile(clientKeyPath)
	if err != nil {
		log.Printf("loading %s: %s\n", clientKeyPath, err)
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Print("initializing X509: %s", err)
		return nil, err
	}

	log.Println("setting up httpclient with TLS")
	log.Printf("API URL: %s\n", po.remoteURL)
	log.Printf("API Ver: %s\n", po.clientVer)
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			},
		},
	}
	return c, nil
}

// instLabel generates the selector used to find the cluster based on instID
func (po *PGOperator) instLabel(instID string) string {
	return po.instLabelKey + "=" + instID
}

// createRequestByPlan updates cluster creation requests associated with plan
// unique identifiers
func (po *PGOperator) createRequestByPlan(planID string, req *msgs.CreateClusterRequest) {
	switch planID {
	case "885a1cb6-ca42-43e9-a725-8195918e1343":
		req.MetricsFlag = true
		req.ContainerResources = "osbsmall"
		req.StorageConfig = "osbsmall"
	case "dc951396-bb28-45a4-b040-cfe3bebc6121":
		req.MetricsFlag = true
		req.ContainerResources = "osbmedium"
		req.StorageConfig = "osbmedium"
	case "04349656-4dc9-4b67-9b15-52a93d64d566":
		req.MetricsFlag = true
		req.ContainerResources = "osblarge"
		req.StorageConfig = "osblarge"
	case "877432f8-07eb-4e57-b984-d025a71d2282":
		req.MetricsFlag = true
		req.ReplicaCount = 1
		req.ContainerResources = "osbsmall"
		req.StorageConfig = "osbsmall"
	case "89bcdf8a-e637-4bb3-b7ce-aca083cc1e69":
		req.MetricsFlag = true
		req.ReplicaCount = 1
		req.ContainerResources = "osbmedium"
		req.StorageConfig = "osbmedium"
	case "470ca1a0-2763-41f1-a4cf-985acdb549ab":
		req.MetricsFlag = true
		req.ReplicaCount = 1
		req.ContainerResources = "osblarge"
		req.StorageConfig = "osblarge"
	default:
		return
	}
}

// CreateBinding creates and/or returns binding information for a cluster
func (po *PGOperator) CreateBinding(instanceID, bindID, appID string) (BasicCred, error) {
	log.Printf("CreateBinding called %s\n", instanceID)
	log.Printf("Binding: %s\n", bindID)
	if appID != "" {
		log.Printf("App ID: %s\n", appID)
	}
	hc, err := po.httpClient()
	if err != nil {
		return BasicCred{}, err
	}

	ns, err := po.findInstanceNamespace(instanceID)
	if err != nil {
		log.Printf("error finding instance in CreateBinding: %s\n", err)
		return BasicCred{}, err
	}

	nu, err := CompactUUIDString(bindID)
	if err != nil {
		return BasicCred{}, fmt.Errorf("unable to process bindID: %s\n", err)
	}
	newUser := fmt.Sprintf("user%s", strings.ToLower(nu))

	cuReq := msgs.CreateUserRequest{
		Username:       newUser,
		Namespace:      ns,
		Selector:       po.instLabel(instanceID),
		ManagedUser:    true,
		ClientVersion:  po.clientVer,
		PasswordLength: 16,
	}
	cuResp, err := api.CreateUser(hc, &po.pgoCreds, &cuReq)
	if err != nil {
		log.Printf("Unable to create user %s: %s\n", newUser, err)
		return BasicCred{}, err
	}
	if cuResp.Code != msgs.Ok {
		log.Printf("Unable to create user %s: %s\n", newUser, cuResp.Msg)
	}

	suReq := &msgs.ShowUserRequest{
		AllFlag:       true,
		ClientVersion: po.clientVer,
		Namespace:     ns,
		Selector:      po.instLabel(instanceID),
	}
	suResp, err := api.ShowUser(hc, &po.pgoCreds, suReq)
	if err != nil {
		log.Printf("error getting user details: %s\n", err)
		return BasicCred{}, err
	}
	if suResp.Status.Code != msgs.Ok {
		m := suResp.Status.Msg
		log.Println(m)
		return BasicCred{}, errors.New("error showing user: " + m)
	}
	if len(suResp.Results) == 0 {
		log.Println("no users found")
		return BasicCred{}, errors.New("no users found for instance " + instanceID)
	}
	users := suResp.Results[0]
	log.Println("cluster secrets are:")
	credentials := make(map[string]interface{})
	for _, s := range users.Secrets {
		if os.Getenv("CRUNCHY_DEBUG") == "true" {
			log.Println("secret : " + s.Name)
			log.Println("username: " + s.Username)
			log.Println("password: " + s.Password)
		}
		credentials[s.Username] = s.Password
	}

	if pass, ok := credentials[newUser]; !ok {
		return BasicCred{}, errors.New("Unable to find newly created user in cluster users")
	} else {
		if pw, ok := pass.(string); ok {
			return BasicCred{Username: newUser, Password: pw}, nil
		} else {
			return BasicCred{}, errors.New("Unrecognized type for password in API response")
		}
	}
}

// ClusterDetail returns the content provided by the operator's Show Cluster
func (po *PGOperator) ClusterDetail(instanceID string) (ClusterDetails, error) {
	log.Printf("ClusterDetail called %s\n", instanceID)
	noInfo := ClusterDetails{}
	hc, err := po.httpClient()
	if err != nil {
		return noInfo, err
	}

	ns, err := po.findInstanceNamespace(instanceID)
	if err != nil {
		log.Printf("error finding instance in ClusterDetails: %s", err)
		return noInfo, err
	}

	showClusterRequest := msgs.ShowClusterRequest{
		Clustername:   "all",
		Selector:      po.instLabel(instanceID),
		ClientVersion: po.clientVer,
		Namespace:     ns,
	}
	response, err := api.ShowCluster(hc, &po.pgoCreds, &showClusterRequest)

	if response.Status.Code == msgs.Ok {
		for _, result := range response.Results {
			log.Println(result)
		}
	} else {
		log.Print(response.Status.Msg)
		return noInfo, errors.New("ShowCluster response: " + response.Status.Msg)
	}

	if len(response.Results) != 1 {
		//error, should always return a single cluster detail
		//because we are using a instanceID as the search key
		return noInfo, errors.New("cluster for instanceID " + instanceID + " not found by ShowCluster")
	}

	detail := &response.Results[0]
	if l := len(detail.Services); l != 1 {
		return noInfo, fmt.Errorf("unexpected number of services for cluster: %d", l)
	}
	svc := detail.Services[0]

	cDetail := ClusterDetails{
		Name:        svc.Name,
		ClusterIP:   svc.ClusterIP,
		ClusterName: svc.ClusterName,
		ExternalIP:  svc.ExternalIP,
	}

	return cDetail, nil
}

// CreateCluster implements the PGOperator interface for creating clusters
func (po *PGOperator) CreateCluster(req CreateRequest) error {
	log.Printf("CreateCluster called %s\n", req.InstanceID)
	hc, err := po.httpClient()
	if err != nil {
		return err
	}

	r := &msgs.CreateClusterRequest{
		ClientVersion: po.clientVer,
		Name:          req.Name,
		Namespace:     req.Namespace,
		Series:        1,
		UserLabels:    po.instLabel(req.InstanceID),
	}
	po.createRequestByPlan(req.PlanID, r)
	log.Println("user label applied to cluster is [" + r.UserLabels + "]")

	log.Printf("creation request: %#v\n", r)
	response, err := api.CreateCluster(hc, &po.pgoCreds, r)
	if err != nil {
		log.Println("create cluster error: ", err)
		return err
	} else if response.Status.Code != msgs.Ok {
		log.Println("create cluster non-Ok status: ", response.Msg)
		return errors.New(response.Msg)
	} else {
		for _, v := range response.Results {
			log.Println(v)
		}
	}

	return nil
}

// DeleteBinding deletes existing binding users based on instance and bindID
func (po *PGOperator) DeleteBinding(instanceID, bindID string) error {
	log.Printf("DeleteBinding called %s\n", instanceID)
	hc, err := po.httpClient()
	if err != nil {
		return err
	}

	ns, err := po.findInstanceNamespace(instanceID)
	if err != nil {
		log.Printf("error finding instance in ClusterDetails: %s", err)
		return err
	}

	u, err := CompactUUIDString(bindID)
	if err != nil {
		return fmt.Errorf("unable to process bindID: %s\n", err)
	}
	user := fmt.Sprintf("user%s", strings.ToLower(u))

	duReq := msgs.DeleteUserRequest{
		AllFlag:       true,
		ClientVersion: po.clientVer,
		Namespace:     ns,
		Selector:      po.instLabel(instanceID),
		Username:      user,
	}
	resp, err := api.DeleteUser(hc, &po.pgoCreds, &duReq)
	if err != nil {
		return err
	}
	if resp.Status.Code == msgs.Ok {
		log.Printf("Deleted user for binding %s\n", bindID)
	} else {
		return fmt.Errorf("response error to delete user: %s", resp.Msg)
	}

	return nil
}

// DeleteCluster implements the PGOperator interface for deleting clusters
// It also ensures all bindings are deleted prior to attempting to delete
// the cluster so that a clear error can be returned
func (po *PGOperator) DeleteCluster(instanceID string) error {
	log.Printf("DeleteCluster called %s\n", instanceID)
	hc, err := po.httpClient()
	if err != nil {
		return err
	}
	selector := po.instLabel(instanceID)

	ns, err := po.findInstanceNamespace(instanceID)
	if err != nil {
		log.Printf("error finding instance in DeleteCluster: %s\n", err)
		return ErrNoInstance{ID: instanceID}
	}

	// Ensure no bound users exist
	suReq := &msgs.ShowUserRequest{
		AllFlag:       true,
		ClientVersion: po.clientVer,
		Namespace:     ns,
		Selector:      po.instLabel(instanceID),
	}
	suResp, err := api.ShowUser(hc, &po.pgoCreds, suReq)
	if err != nil {
		log.Printf("error getting user details: %s\n", err)
		return err
	}
	if suResp.Status.Code != msgs.Ok {
		m := suResp.Status.Msg
		log.Println(m)
		return errors.New("error fetching users: " + m)
	}
	if len(suResp.Results) == 0 {
		log.Println("no users found, expected default users")
		return errors.New("unexpected user state: no default users " + instanceID)
	}
	users := suResp.Results[0]
	for _, s := range users.Secrets {
		if strings.Compare("user", s.Username[:4]) == 0 {
			return ErrBindingsRemain
		}
	}

	// Proceed with deletion
	deleteData := false
	deleteBackups := false
	log.Printf("deleting cluster %s with delete-data %t\n", selector, deleteData)

	deleteClusterRequest := msgs.DeleteClusterRequest{
		Clustername:   "all",
		Selector:      selector,
		ClientVersion: po.clientVer,
		Namespace:     ns,
		DeleteData:    deleteData,
		DeleteBackups: deleteBackups,
	}
	response, err := api.DeleteCluster(hc, &deleteClusterRequest, &po.pgoCreds)

	if response.Status.Code == msgs.Ok {
		for _, result := range response.Results {
			log.Println(result)
		}
	} else {
		log.Print(response.Status.Msg)
	}

	return err
}
