#!/bin/bash
# Copyright 2017-2018 Crunchy Data Solutions, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

$DIR/cleanup.sh

if [ "$CO_CMD" = "kubectl" ]; then
	NS="--namespace=$CO_NAMESPACE"
fi


$CO_CMD $NS create secret tls pgo-osb-tls-secret --cert=$DIR/server.crt --key=$DIR/server.key -o yaml

$CO_CMD $NS create secret generic pgo-osb-apiserver-secret \
        --from-file=ca=$DIR/server.crt \
        --from-file=clientcert=$DIR/server.crt \
        --from-file=clientkey=$DIR/server.key

expenv -f $DIR/service-account.yaml | $CO_CMD create -f -
expenv -f $DIR/secret.yaml | $CO_CMD create -f -

$CO_CMD create -f $DIR/cluster-role-1.yaml
$CO_CMD create -f $DIR/cluster-role.yaml
expenv -f $DIR/cluster-role-binding-1.yaml | $CO_CMD  create -f -
expenv -f $DIR/cluster-role-binding.yaml | $CO_CMD  create -f -
$CO_CMD  create -f $DIR/service.yaml

expenv -f $DIR/deployment.yaml | $CO_CMD  create -f -

echo "sleeping before we create clusterservicebroker"
sleep 15

expenv -f $DIR/cluster-service-broker-no-auth.yaml | $CO_CMD  create -f -

