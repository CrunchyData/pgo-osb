#!/bin/bash

$CO_CMD delete sa pgo-osb
expenv -f $DIR/service-account.yaml | $CO_CMD create -f -

$CO_CMD delete secret pgo-osb
expenv -f $DIR/secret.yaml | $CO_CMD create -f -

$CO_CMD delete clusterservicebroker pgo-osb
expenv -f $DIR/cluster-service-broker.yaml | $CO_CMD  create -f -

$CO_CMD delete clusterrole pgo-osb access-pgo-osb

$CO_CMD delete clusterrolebinding pgo-osb pgo-osb-client

$CO_CMD delete service pgo-osb

$CO_CMD delete deployment pgo-osb

