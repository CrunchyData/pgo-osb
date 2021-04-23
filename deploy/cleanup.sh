#!/bin/bash 
# Copyright 2017-2021 Crunchy Data Solutions, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICEOSB_NAMESPACEE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIOOSB_NAMESPACE OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


#if [ "$OSB_CMD" = "kubectl" ]; then
#	OSB_NAMESPACE="--namespace=$OSB_NAMESPACE"
#fi

$OSB_CMD --namespace=$OSB_NAMESPACE delete secret pgo-osb pgo-osb-apiserver-secret pgo-osb-tls-secret
$OSB_CMD --namespace=$OSB_NAMESPACE delete serviceaccount pgo-osb

$OSB_CMD --namespace=$OSB_NAMESPACE delete clusterrolebinding pgo-osb pgo-osb-client
$OSB_CMD --namespace=$OSB_NAMESPACE delete clusterrole pgo-osb access-pgo-osb

$OSB_CMD --namespace=$OSB_NAMESPACE delete service pgo-osb

$OSB_CMD --namespace=$OSB_NAMESPACE delete deployment pgo-osb

$OSB_CMD --namespace=$OSB_NAMESPACE delete clusterservicebroker pgo-osb

sleep 5

