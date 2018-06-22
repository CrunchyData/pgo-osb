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
	log "github.com/Sirupsen/logrus"
	msgs "github.com/crunchydata/pgo-osb/apiservermsgs"
	"io/ioutil"
	"net/http"
	"strconv"
)

// DeleteCluster ...
func DeleteCluster(APIServerURL, basicAuthUsername, basicAuthPassword, clusterName, clientVersion string, deleteData, deleteBackups bool) error {
	log.Debugf("deleteCluster called %s\n", clusterName)
	selector := "name=" + clusterName

	log.Debug("deleting cluster " + clusterName + " with delete-data " + strconv.FormatBool(deleteData))

	url := APIServerURL + "/clustersdelete/" + clusterName + "?selector=" + selector + "&delete-data=" + strconv.FormatBool(deleteData) + "&delete-backups=" + strconv.FormatBool(deleteBackups) + "&version=" + clientVersion

	log.Debug("delete cluster called [" + url + "]")

	action := "GET"
	req, err := http.NewRequest(action, url, nil)
	if err != nil {
		log.Error("NewRequest: ", err)
		return err
	}

	req.SetBasicAuth(basicAuthUsername, basicAuthPassword)

	httpclient, err := GetCredentials(basicAuthUsername, basicAuthPassword)
	if err != nil {
		return err
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Error("Do: ", err)
		return err
	}
	log.Debugf("%v\n", resp)
	if !StatusCheck(resp) {
		return errors.New("could not authenticate")
	}

	defer resp.Body.Close()
	var response msgs.DeleteClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Error(err)
		log.Println(err)
		return err
	}

	if response.Status.Code == msgs.Ok {
		for _, result := range response.Results {
			log.Infoln(result)
		}
	} else {
		log.Error(response.Status.Msg)
	}

	return err

}

// CreateCluster ....
func CreateCluster(APIServerURL, BasicAuthUsername, BasicAuthPassword, clusterName, clientVersion string) error {
	var err error

	r := new(msgs.CreateClusterRequest)
	r.Name = clusterName
	//r.NodeLabel = NodeLabel
	//r.Password = Password
	//r.SecretFrom = SecretFrom
	//r.BackupPVC = BackupPVC
	//r.UserLabels = UserLabels
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
	log.Debug("createCluster called...[" + url + "]")

	action := "POST"
	req, err := http.NewRequest(action, url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("NewRequest: ", err)
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
		log.Error("Do: ", err)
		return err
	}

	log.Debugf("%v\n", resp)
	if !StatusCheck(resp) {
		return errors.New("could not authenticate")
	}

	defer resp.Body.Close()

	var response msgs.CreateClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%v\n", resp.Body)
		log.Error(err)
		log.Println(err)
		return err
	}

	if response.Status.Code == msgs.Ok {
		for _, v := range response.Results {
			log.Infoln(v)
		}
	} else {
		log.Error(response.Status.Msg)
	}

	return err

}

// StatusCheck ...
func StatusCheck(resp *http.Response) bool {
	log.Debugf("http status code is %d\n", resp.StatusCode)
	if resp.StatusCode == 401 {
		log.Error("Authentication Failed: %d\n", resp.StatusCode)
		return false
	} else if resp.StatusCode != 200 {
		log.Error("Invalid Status Code: %d\n", resp.StatusCode)
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
		log.Error(err)
		log.Error(caCertPath + " not found")
		return httpclient, err
	}
	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCertPath = "/opt/apiserver-keys/client.crt"
	_, err = ioutil.ReadFile(clientCertPath)
	if err != nil {
		log.Error(clientCertPath + " not found")
		return httpclient, err
	}

	clientKeyPath = "/opt/apiserver-keys/client.key"
	_, err = ioutil.ReadFile(clientKeyPath)
	if err != nil {
		log.Error(clientKeyPath + " not found")
		return httpclient, err
	}

	cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Error(err)
		return httpclient, err
	}

	log.Debug("setting up httpclient with TLS")
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
