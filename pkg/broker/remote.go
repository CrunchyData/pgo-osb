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
	"k8s.io/client-go/rest"
)

// RESTClient represents the REST client for Kube
// TODO: Can RESTClient be easily pulled into remote, logically belongs
var RESTClient *rest.RESTClient

// BasicCred represents a common pair of username and password
type BasicCred struct {
	Username string
	Password string
}

type ClusterDetails struct {
	Name        string
	ClusterIP   string
	ExternalIP  string
	ClusterName string
}

// Remote defines an interface for servicing OSB requests
type Remote interface {
	BindingUser(instanceID, appID, bindID string) (BasicCred, error)
	ClusterDetail(instanceID string) (ClusterDetails, error)
	CreateCluster(instanceID, name, namespace string) error
	DeleteBinding(instanceID, bindID string) error
	DeleteCluster(instanceID string) error
}
