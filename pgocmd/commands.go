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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	msgs "github.com/crunchydata/postgres-operator/apiservermsgs"
	//"github.com/crunchydata/postgres-operator/pgo/cmd"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const INSTANCE_LABEL_KEY = "pgo-osb-instance"

// GetClusterCredentials ...
func GetClusterCredentials(APIServerURL, basicAuthUsername, basicAuthPassword, clientVersion, instanceID string) (map[string]interface{}, []msgs.ShowClusterService, error) {

	credentials := make(map[string]interface{})
	var response *msgs.ShowClusterResponse
	var detail *msgs.ShowClusterDetail
	log.Printf("ShowCluster called %s\n", instanceID)
	selector := INSTANCE_LABEL_KEY + "=" + instanceID

	clusterName := "all"
	log.Print("show cluster " + selector)

	url := APIServerURL + "/clusters/" + clusterName + "?selector=" + selector + "&version=" + clientVersion

	log.Print("show cluster called [" + url + "]")

	action := "GET"
	req, err := http.NewRequest(action, url, nil)
	if err != nil {
		log.Print("NewRequest: ", err)
		return credentials, nil, err
	}

	req.SetBasicAuth(basicAuthUsername, basicAuthPassword)

	httpclient, err := GetCredentials(basicAuthUsername, basicAuthPassword)
	if err != nil {
		return credentials, nil, err
	}

	/**
	args := make([]string, 1)
	args[0] = "all"
	cmd.Selector = INSTANCE_LABEL_KEY + "=" + instanceID
	cmd.Httpclient = httpclient
	log.Println("here is jeff pgo cmd call")
	cmd.ShowCluster(args)
	log.Println("aster jeff pgo cmd call")
	*/

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Print("Do: ", err)
		return credentials, nil, err
	}
	log.Printf("%v\n", resp)
	if !StatusCheck(resp) {
		return credentials, nil, errors.New("could not authenticate")
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Print(err)
		log.Println(err)
		return credentials, nil, err
	}

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
	users := showUser(basicAuthUsername, basicAuthPassword, APIServerURL, clientVersion, clusterName, selector)

	fmt.Println("cluster secrets are:")
	//for _, s := range detail.Secrets {
	for _, s := range users.Secrets {
		fmt.Println("secret : " + s.Name)
		fmt.Println("username: " + s.Username)
		fmt.Println("password: " + s.Password)
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

	url := APIServerURL + "/clustersdelete/" + clusterName + "?selector=" + selector + "&delete-data=" + strconv.FormatBool(deleteData) + "&delete-backups=" + strconv.FormatBool(deleteBackups) + "&version=" + clientVersion

	log.Print("delete cluster called [" + url + "]")

	action := "GET"
	req, err := http.NewRequest(action, url, nil)
	if err != nil {
		log.Print("NewRequest: ", err)
		return err
	}

	req.SetBasicAuth(basicAuthUsername, basicAuthPassword)

	httpclient, err := GetCredentials(basicAuthUsername, basicAuthPassword)
	if err != nil {
		return err
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Print("Do: ", err)
		return err
	}
	log.Printf("%v\n", resp)
	if !StatusCheck(resp) {
		return errors.New("could not authenticate")
	}

	defer resp.Body.Close()
	var response msgs.DeleteClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Print(err)
		log.Println(err)
		return err
	}

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
func CreateCluster(APIServerURL, BasicAuthUsername, BasicAuthPassword, clusterName, clientVersion, instanceID string) error {
	var err error

	r := new(msgs.CreateClusterRequest)
	r.Name = clusterName
	//r.NodeLabel = NodeLabel
	//r.Password = Password
	//r.SecretFrom = SecretFrom
	//r.BackupPVC = BackupPVC
	r.UserLabels = INSTANCE_LABEL_KEY + "=" + instanceID
	fmt.Println("user label applied is [" + r.UserLabels + "]")
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

	jsonValue, _ := json.Marshal(r)
	url := APIServerURL + "/clusters"
	log.Print("createCluster called...[" + url + "]")

	action := "POST"
	req, err := http.NewRequest(action, url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Print("NewRequest: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(BasicAuthUsername, BasicAuthPassword)

	httpclient, err := GetCredentials(BasicAuthUsername, BasicAuthPassword)
	if err != nil {
		return err
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Print("Do: ", err)
		return err
	}

	log.Printf("%v\n", resp)
	if !StatusCheck(resp) {
		return errors.New("could not authenticate")
	}

	defer resp.Body.Close()

	var response msgs.CreateClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Print(err)
		log.Println(err)
		return err
	}

	if response.Status.Code == msgs.Ok {
		for _, v := range response.Results {
			log.Println(v)
		}
	} else {
		log.Print(response.Status.Msg)
	}

	return err

}

// StatusCheck ...
func StatusCheck(resp *http.Response) bool {
	log.Printf("http status code is %d\n", resp.StatusCode)
	if resp.StatusCode == 401 {
		log.Print("Authentication Failed: %d\n", resp.StatusCode)
		return false
	} else if resp.StatusCode != 200 {
		log.Print("Invalid Status Code: %d\n", resp.StatusCode)
		return false
	}

	return true
}

func GetCredentials(username, password string) (*http.Client, error) {
	var err error
	var httpclient *http.Client

	var caCertPool *x509.CertPool
	var cert tls.Certificate
	var caCertPath, clientCertPath, clientKeyPath string

	caCertPath = "/opt/apiserver-keys/ca.crt"
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Print(err)
		log.Print(caCertPath + " not found")
		return httpclient, err
	}
	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCertPath = "/opt/apiserver-keys/client.crt"
	_, err = ioutil.ReadFile(clientCertPath)
	if err != nil {
		log.Print(clientCertPath + " not found")
		return httpclient, err
	}

	clientKeyPath = "/opt/apiserver-keys/client.key"
	_, err = ioutil.ReadFile(clientKeyPath)
	if err != nil {
		log.Print(clientKeyPath + " not found")
		return httpclient, err
	}

	cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Print(err)
		return httpclient, err
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

	return httpclient, err
}

// showUser ...
func showUser(BasicAuthUsername, BasicAuthPassword, APIServerURL, clientVersion, clusterName, selector string) msgs.ShowUserDetail {

	var userDetail msgs.ShowUserDetail

	log.Print("showUser called %v\n", clusterName)

	url := APIServerURL + "/users/" + clusterName + "?selector=" + selector + "&version=" + clientVersion

	log.Print("show users called [" + url + "]")

	action := "GET"
	req, err := http.NewRequest(action, url, nil)

	if err != nil {
		log.Printf("NewRequest: %v ", err)
		return userDetail
	}

	req.SetBasicAuth(BasicAuthUsername, BasicAuthPassword)

	httpclient, err := GetCredentials(BasicAuthUsername, BasicAuthPassword)
	if err != nil {
		return userDetail
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Printf("Do: %v", err)
		return userDetail
	}
	log.Printf("%v\n", resp)
	StatusCheck(resp)

	defer resp.Body.Close()

	var response msgs.ShowUserResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Println(err)
		log.Println(err)
		return userDetail
	}

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
