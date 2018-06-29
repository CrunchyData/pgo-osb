
PULL ?= IfNotPresent

build: ## Builds the starter pack
	go build -i github.com/crunchydata/pgo-osb/cmd/pgo-osb

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o pgo-osb --ldflags="-s" github.com/crunchydata/pgo-osb/cmd/pgo-osb

image: linux ## Builds a Linux based image
	docker build -t pgo-osb -f $(CO_BASEOS)/Dockerfile.pgo-osb.$(CO_BASEOS) .
	docker tag pgo-osb $(CO_IMAGE_PREFIX)/pgo-osb:$(CO_IMAGE_TAG)

deploy:
	cd deploy && ./deploy.sh

clean: ## Cleans up build artifacts
	rm -f pgo-osb
	rm -f pgo-osb-linux
	rm -f image/pgo-osb

provision: ## Provisions a service instance
	kubectl apply -f manifests/service-instance.yaml

setup: 
	go get github.com/blang/expenv

bind: ## Creates a binding
	kubectl apply -f manifests/service-binding.yaml

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: deploy build test linux image clean push deploy-helm deploy-openshift create-ns provision bind help
