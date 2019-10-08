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

import ()

// BasicCred represents a common pair of username and password
type BasicCred struct {
	Username string
	Password string
}

// ClusterDetails encapsulates information returned about the cluster
type ClusterDetails struct {
	Name        string
	ClusterIP   string
	ExternalIP  string
	ClusterName string
}

type CreateRequest struct {
	InstanceID string
	Name       string
	Namespace  string
	PlanID     string
}

// Executor defines an interface for servicing OSB requests
type Executor interface {
	Provisioner
	Binder
	ClusterDetail(instanceID string) (ClusterDetails, error)
}

// Provisioner defines an interface for (de)provisioning clusters
type Provisioner interface {
	CreateCluster(req CreateRequest) error
	DeleteCluster(instanceID string) error
}

// Binder defines an interface for creating and deleting user bindings
type Binder interface {
	CreateBinding(instanceID, bindID, appID string) (BasicCred, error)
	DeleteBinding(instanceID, bindID string) error
}
