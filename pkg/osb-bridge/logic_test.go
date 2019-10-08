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
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/crunchydata/pgo-osb/pkg/broker"

	"github.com/gofrs/uuid"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	osblib "github.com/pmorie/osb-broker-lib/pkg/broker"
)

func mockLogic(t *testing.T) *BusinessLogic {
	bl, err := NewBusinessLogic(Options{
		Simulated: true,
	})
	if err != nil {
		t.Fatalf("error creating BusinessLogic: %s", err)
	}
	return bl
}

func nuuid(t *testing.T) string {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("failed to generate UUID: %s\n", err)
	}
	return id.String()
}

func TestUnitCatalog(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	req := &osblib.RequestContext{}
	resp, err := bl.GetCatalog(req)
	if err != nil {
		t.Errorf("error getting catalog: %s\n", err)
	}

	if l := len(resp.Services); l != 1 {
		t.Fatalf("expected one and only one service, found %d", l)
	}

	svc := resp.Services[0]

	if v := svc.Name; v != "pgo-osb-service" {
		t.Errorf("unexpected service name: %s", v)
	}

	if svc.Bindable != true {
		t.Error("expected service definition to be bindable")
	}

	if svc.PlanUpdatable != nil && *svc.PlanUpdatable == true {
		t.Errorf("Update function unimplemented, expected PlanUpdatable to be false or undefined")
	}

	if l := len(svc.Plans); l != 7 {
		t.Fatalf("expected seven plans, found %d", l)
	}

	// Some platforms (PCF) freak out if the plan name changes or goes away
	// Do not blindly update this test case without taking that into account
	for _, plan := range svc.Plans {
		switch plan.Name {
		case "default":
			if plan.ID != "86064792-7ea2-467b-af93-ac9694d96d5c" {
				t.Error("unexpected plan Name or ID change for default plan")
			}
		case "standalone_sm":
			if plan.ID != "885a1cb6-ca42-43e9-a725-8195918e1343" {
				t.Error("unexpected plan Name or ID change for standalone_sm plan")
			}
		case "standalone_md":
			if plan.ID != "dc951396-bb28-45a4-b040-cfe3bebc6121" {
				t.Error("unexpected plan Name or ID change for standalone_md plan")
			}
		case "standalone_lg":
			if plan.ID != "04349656-4dc9-4b67-9b15-52a93d64d566" {
				t.Error("unexpected plan Name or ID change for standalone_lg plan")
			}
		case "ha_sm":
			if plan.ID != "877432f8-07eb-4e57-b984-d025a71d2282" {
				t.Error("unexpected plan Name or ID change for ha_sm plan")
			}
		case "ha_md":
			if plan.ID != "89bcdf8a-e637-4bb3-b7ce-aca083cc1e69" {
				t.Error("unexpected plan Name or ID change for ha_md plan")
			}
		case "ha_lg":
			if plan.ID != "470ca1a0-2763-41f1-a4cf-985acdb549ab" {
				t.Error("unexpected plan Name or ID change for ha_lg plan")
			}
		default:
			t.Errorf("Unexpected plan name: %s", plan.Name)
		}
	}
}

func TestUnitProvisionBasic(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	req := &osb.ProvisionRequest{
		InstanceID: nuuid(t),
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
		Parameters: map[string]interface{}{
			"PGO_NAMESPACE":   "demo",
			"PGO_CLUSTERNAME": "unitinstance",
		},
	}

	_, err := bl.Provision(req, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitProvisionMissingClustername(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	req := &osb.ProvisionRequest{
		InstanceID: nuuid(t),
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
		Parameters: map[string]interface{}{
			"PGO_NAMESPACE": "demo",
		},
	}

	_, err := bl.Provision(req, nil)
	if err == nil {
		t.Fatal("expected provisioning error re: Namespace, but got none")
	}
}

func TestUnitProvisionMissingNamespace(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	req := &osb.ProvisionRequest{
		InstanceID: nuuid(t),
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
		Parameters: map[string]interface{}{
			"PGO_CLUSTERNAME": "unitinstance",
		},
	}

	_, err := bl.Provision(req, nil)
	if err == nil {
		t.Fatal("expected provisioning error re: Clustername, but got none")
	}
}

func TestUnitProvisionUndo(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	preq := &osb.ProvisionRequest{
		InstanceID: nuuid(t),
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
		Parameters: map[string]interface{}{
			"PGO_NAMESPACE":   "demo",
			"PGO_CLUSTERNAME": "unitinstance",
		},
	}

	_, err := bl.Provision(preq, nil)
	if err != nil {
		t.Fatalf("error provisioning: %s", err)
	}

	dreq := &osb.DeprovisionRequest{
		InstanceID: preq.InstanceID,
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
	}
	_, err = bl.Deprovision(dreq, nil)
	if err != nil {
		t.Fatalf("error deprovisioning: %s", err)
	}
}

func TestUnitBindingNoInstance(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	preq := &osb.ProvisionRequest{
		InstanceID: nuuid(t),
		PlanID:     "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:  "4be12541-2945-4101-8a33-79ac0ad58750",
		Parameters: map[string]interface{}{
			"PGO_NAMESPACE":   "demo",
			"PGO_CLUSTERNAME": "unitinstance",
		},
	}

	_, err := bl.Provision(preq, nil)
	if err != nil {
		t.Fatalf("error provisioning: %s", err)
	}

	appID := nuuid(t)
	breq := &osb.BindRequest{
		InstanceID: nuuid(t),
		BindingID:  nuuid(t),
		AppGUID:    &appID,
	}

	_, err = bl.Bind(breq, nil)
	if err == nil {
		t.Fatal("NoInstance error expected, got nil")
	}
	if _, ok := err.(broker.ErrNoInstance); !ok {
		t.Fatalf("NoInstance error expected, got: %T - %s", err, err)
	}
}

func TestUnitBindingBasic(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	bl := mockLogic(t)
	preq := &osb.ProvisionRequest{
		InstanceID:        nuuid(t),
		PlanID:            "86064792-7ea2-467b-af93-ac9694d96d5c",
		ServiceID:         "4be12541-2945-4101-8a33-79ac0ad58750",
		AcceptsIncomplete: false,
		Parameters: map[string]interface{}{
			"PGO_NAMESPACE":   "demo",
			"PGO_CLUSTERNAME": "unitinstance",
		},
	}

	_, err := bl.Provision(preq, nil)
	if err != nil {
		t.Fatalf("error provisioning: %s", err)
	}

	appID := nuuid(t)
	breq := &osb.BindRequest{
		InstanceID: preq.InstanceID,
		BindingID:  nuuid(t),
		AppGUID:    &appID,
	}

	bindResp, err := bl.Bind(breq, nil)
	if err != nil {
		t.Fatalf("error binding: %s", err)
	}

	h := md5.New()
	io.WriteString(h, breq.BindingID)
	expUser := fmt.Sprintf("user_%x", h.Sum(nil))
	expect := &osblib.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{
				"username":      expUser,
				"password":      broker.MockStatic.Password,
				"db_port":       5432,
				"db_name":       "userdb",
				"db_host":       broker.MockStatic.ExternalIP,
				"internal_host": broker.MockStatic.ClusterIP,
				"uri": fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
					expUser,
					broker.MockStatic.Password,
					broker.MockStatic.ExternalIP,
					5432,
					"userdb"),
			},
		},
	}

	if !reflect.DeepEqual(expect, bindResp) {
		t.Logf("Expected: %+v", expect)
		t.Logf("Received: %+v", bindResp)
		t.FailNow()
	}
}
