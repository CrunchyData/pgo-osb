#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export BROKER_CA_CERT=
export IMAGE=crunchydata/pgo-osb:centos7-0.0.1

# kubectl create secret tls jeff-broker-tls-secret --cert=/tmp/server.crt --key=/tmp/server.key -o yaml

kubectl create secret generic pgo-osb-apiserver-secret \
	--from-file=ca=$DIR/server.crt \
	--from-file=clientcert=$DIR/server.crt \
	--from-file=clientkey=$DIR/server.key 

expenv -f $DIR/service-account.yaml | $CO_CMD create -f -
expenv -f $DIR/secret.yaml | $CO_CMD create -f -
#expenv -f $DIR/service-account-2.yaml | $CO_CMD  create -f -
$CO_CMD create -f $DIR/cluster-role-1.yaml
$CO_CMD create -f $DIR/cluster-role.yaml 
expenv -f $DIR/cluster-role-binding-1.yaml | $CO_CMD  create -f -
expenv -f $DIR/cluster-role-binding.yaml | $CO_CMD  create -f -
$CO_CMD  create -f $DIR/service.yaml 
expenv -f $DIR/deployment.yaml | $CO_CMD  create -f -
echo "sleeping before we create clusterservicebroker"
sleep 15
expenv -f $DIR/cluster-service-broker-no-auth.yaml | $CO_CMD  create -f -

