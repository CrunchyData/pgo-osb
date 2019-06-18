package pgocmd

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
	"strconv"

	crv1 "github.com/crunchydata/postgres-operator/apis/cr/v1"
	msgs "github.com/crunchydata/postgres-operator/apiservermsgs"
	"github.com/crunchydata/postgres-operator/kubeapi"
	api "github.com/crunchydata/postgres-operator/pgo/api"
	"k8s.io/client-go/rest"
)

const INSTANCE_LABEL_KEY = "pgo-osb-instance"

// RESTClient ...
var RESTClient *rest.RESTClient

// GetClusterCredentials ...
func GetClusterCredentials(APIServerURL, basicAuthUsername, basicAuthPassword, clientVersion, instanceID string) (map[string]interface{}, []msgs.ShowClusterService, error) {

	credentials := make(map[string]interface{})
	var detail *msgs.ShowClusterDetail
	fmt.Printf("fmt ShowCluster called %s\n", instanceID)
	log.Printf("ShowCluster called %s\n", instanceID)
	selector := INSTANCE_LABEL_KEY + "=" + instanceID

	clusterName := "all"
	log.Print("show cluster " + selector)
	httpclient, SessionCredentials, err := GetCredentials(basicAuthUsername, basicAuthPassword, APIServerURL)
	if err != nil {
		return credentials, nil, err
	}

	clusterList := crv1.PgclusterList{}
	err = kubeapi.GetpgclustersBySelector(RESTClient, &clusterList, selector, "")
	if err != nil {
		return nil, nil, err
	}
	pgCluster := clusterList.Items[0]

	ccpImageTag := ""
	showClusterRequest := msgs.ShowClusterRequest{
		Clustername:   clusterName,
		Selector:      selector,
		Ccpimagetag:   ccpImageTag,
		ClientVersion: clientVersion,
		Namespace:     pgCluster.GetNamespace(),
	}
	response, err := api.ShowCluster(httpclient, SessionCredentials, &showClusterRequest)

	if response.Status.Code == msgs.Ok {
		for _, result := range response.Results {
			log.Println(result)
		}
	} else {
		log.Print(response.Status.Msg)
		return credentials, nil, err
	}

	if len(response.Results) != 1 {
		//error, should always return a single cluster detail
		//because we are using a instanceID as the search key
		return credentials, nil, errors.New("cluster for instanceID " + instanceID + " not found in bind ")
	}

	detail = &response.Results[0]

	users := showUser(basicAuthUsername, basicAuthPassword, APIServerURL, clientVersion, clusterName, selector,
		pgCluster.GetNamespace())

	log.Println("cluster secrets are:")
	for _, s := range users.Secrets {
		if os.Getenv("CRUNCHY_DEBUG") == "true" {
			log.Println("secret : " + s.Name)
			log.Println("username: " + s.Username)
			log.Println("password: " + s.Password)
		}
		credentials[s.Username] = s.Password
	}

	return credentials, detail.Services, err

}

// DeleteCluster ...
func DeleteCluster(APIServerURL, basicAuthUsername, basicAuthPassword, clientVersion, instanceID string) error {
	log.Printf("deleteCluster called %s\n", instanceID)
	selector := INSTANCE_LABEL_KEY + "=" + instanceID

	clusterName := "all"
	deleteData := false
	deleteBackups := false
	log.Print("deleting cluster " + selector + " with delete-data " + strconv.FormatBool(deleteData))

	httpclient, SessionCredentials, err := GetCredentials(basicAuthUsername, basicAuthPassword, APIServerURL)
	if err != nil {
		return err
	}

	clusterList := crv1.PgclusterList{}
	err = kubeapi.GetpgclustersBySelector(RESTClient, &clusterList, selector, "")
	if err != nil {
		return err
	}
	pgCluster := clusterList.Items[0]

	deleteClusterRequest := msgs.DeleteClusterRequest{
		Clustername:   clusterName,
		Selector:      selector,
		ClientVersion: clientVersion,
		Namespace:     pgCluster.GetNamespace(),
		DeleteData:    deleteData,
		DeleteBackups: deleteBackups,
	}
	response, err := api.DeleteCluster(httpclient, &deleteClusterRequest, SessionCredentials)

	if response.Status.Code == msgs.Ok {
		for _, result := range response.Results {
			log.Println(result)
		}
	} else {
		log.Print(response.Status.Msg)
	}

	return err

}

// CreateCluster ....
func CreateCluster(APIServerURL, BasicAuthUsername, BasicAuthPassword, clusterName, clientVersion, instanceID, namespace string) error {
	var err error

	r := new(msgs.CreateClusterRequest)
	r.Name = clusterName
	//r.NodeLabel = NodeLabel
	//r.Password = Password
	//r.SecretFrom = SecretFrom
	//r.BackupPVC = BackupPVC
	r.UserLabels = INSTANCE_LABEL_KEY + "=" + instanceID
	log.Println("user label applied is [" + r.UserLabels + "]")
	//r.BackupPath = BackupPath
	//r.Policies = PoliciesFlag
	//r.CCPImageTag = CCPImageTag
	r.Series = 1
	//r.MetricsFlag = MetricsFlag
	//r.AutofailFlag = AutofailFlag
	//r.PgpoolFlag = PgpoolFlag
	//r.ArchiveFlag = ArchiveFlag
	//r.PgpoolSecret = PgpoolSecret
	//r.CustomConfig = CustomConfig
	//r.StorageConfig = StorageConfig
	//r.ReplicaStorageConfig = ReplicaStorageConfig
	//r.ContainerResources = ContainerResources
	r.ClientVersion = clientVersion
	r.Namespace = namespace

	httpclient, SessionCredentials, err := GetCredentials(BasicAuthUsername, BasicAuthPassword, APIServerURL)
	if err != nil {
		return err
	}

	response, err := api.CreateCluster(httpclient, SessionCredentials, r)
	if err != nil {
		log.Println(err)
		return err
	} else if response.Status.Code != msgs.Ok {
		log.Println(response.Msg)
		return errors.New(response.Msg)
	} else {
		for _, v := range response.Results {
			log.Println(v)
		}
	}

	return err

}

func GetCredentials(username, password, APIServerURL string) (*http.Client, *msgs.BasicAuthCredentials, error) {
	var err error
	var httpclient *http.Client
	var creds *msgs.BasicAuthCredentials

	var caCertPool *x509.CertPool
	var cert tls.Certificate
	var caCertPath, clientCertPath, clientKeyPath string

	caCertPath = "/opt/apiserver-keys/ca.crt"
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Print(err)
		log.Print(caCertPath + " not found")
		return httpclient, creds, err
	}
	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCertPath = "/opt/apiserver-keys/client.crt"
	_, err = ioutil.ReadFile(clientCertPath)
	if err != nil {
		log.Print(clientCertPath + " not found")
		return httpclient, creds, err
	}

	clientKeyPath = "/opt/apiserver-keys/client.key"
	_, err = ioutil.ReadFile(clientKeyPath)
	if err != nil {
		log.Print(clientKeyPath + " not found")
		return httpclient, creds, err
	}

	cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Print(err)
		return httpclient, creds, err
	}

	log.Println("setting up httpclient with TLS")
	httpclient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			},
		},
	}
	creds = &msgs.BasicAuthCredentials{
		Username:     username,
		Password:     password,
		APIServerURL: APIServerURL,
	}

	return httpclient, creds, err
}

// showUser ...
func showUser(BasicAuthUsername, BasicAuthPassword, APIServerURL, clientVersion, clusterName, selector, ns string) msgs.ShowUserDetail {

	var userDetail msgs.ShowUserDetail

	log.Print("showUser called %v\n", clusterName)
	expired := ""
	httpclient, SessionCredentials, err := GetCredentials(BasicAuthUsername, BasicAuthPassword, APIServerURL)
	if err != nil {
		return userDetail
	}
	response, err := api.ShowUser(httpclient, clusterName, selector, expired, SessionCredentials, ns)

	if response.Status.Code != msgs.Ok {
		log.Println(response.Status.Msg)
		return userDetail
	}
	if len(response.Results) == 0 {
		log.Println("no clusters found")
		return userDetail
	}

	return response.Results[0]

}
