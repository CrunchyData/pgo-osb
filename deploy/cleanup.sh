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


if [ "$CO_CMD" = "kubectl" ]; then
	NS="--namespace=$CO_NAMESPACE"
fi

$CO_CMD $NS delete secret pgo-osb pgo-osb-apiserver-secret pgo-osb-tls-secret
$CO_CMD $NS delete serviceaccount pgo-osb

$CO_CMD $NS delete clusterrolebinding pgo-osb pgo-osb-client
$CO_CMD $NS delete clusterrole pgo-osb access-pgo-osb

$CO_CMD $NS delete service pgo-osb

$CO_CMD $NS delete deployment pgo-osb

$CO_CMD $NS delete clusterservicebroker pgo-osb

sleep 5

