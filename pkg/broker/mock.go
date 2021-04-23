package broker

/*
 Copyright 2017-2021 Crunchy Data Solutions, Inc.
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
	"crypto/md5"
	"fmt"
	"io"
	"sync"
)

// Encapsulate test global data
var MockStatic struct {
	ExternalIP string
	ClusterIP  string
	Password   string
}

func init() {
	MockStatic.ExternalIP = "198.51.100.42" // RFC5737 TEST-NET-2
	MockStatic.ClusterIP = "10.10.33.44"
	MockStatic.Password = "WaltSentMe"
}

type Mock struct {
	sync.RWMutex
	instances map[string]ClusterDetails
	bindings  map[string]BasicCred
}

func NewMock() *Mock {
	m := &Mock{
		instances: map[string]ClusterDetails{},
		bindings:  map[string]BasicCred{},
	}
	return m
}

func (m *Mock) ClusterDetail(instanceID string) (ClusterDetails, error) {
	m.RLock()
	defer m.RUnlock()

	inst, ok := m.instances[instanceID]
	if !ok {
		return ClusterDetails{}, ErrNoInstance{instanceID}
	}

	return inst, nil
}

func (m *Mock) CreateCluster(req CreateRequest) error {
	m.Lock()
	defer m.Unlock()

	m.instances[req.InstanceID] = ClusterDetails{
		Name:        req.Name,
		ClusterName: req.Name,
		ExternalIP:  MockStatic.ExternalIP,
		ClusterIP:   MockStatic.ClusterIP,
	}

	return nil
}

func (m *Mock) DeleteCluster(instanceID string) error {
	m.Lock()
	defer m.Unlock()

	if len(m.bindings) > 0 {
		return ErrBindingsRemain
	}

	delete(m.instances, instanceID)

	return nil
}

func (m *Mock) CreateBinding(instanceID, bindID, appID string) (BasicCred, error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.instances[instanceID]; !ok {
		return BasicCred{}, ErrNoInstance{instanceID}
	}

	key := fmt.Sprintf("%s:%s", instanceID, bindID)
	h := md5.New()
	io.WriteString(h, bindID)
	user := fmt.Sprintf("user_%x", h.Sum(nil))
	m.bindings[key] = BasicCred{
		Username: user,
		Password: MockStatic.Password,
	}

	return m.bindings[key], nil
}

func (m *Mock) DeleteBinding(instanceID, bindID string) error {
	m.Lock()
	defer m.Unlock()

	key := fmt.Sprintf("%s:%s", instanceID, bindID)
	delete(m.bindings, key)

	return nil
}
