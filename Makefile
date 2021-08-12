# Default values if not already set
OSB_BASEOS ?= centos8
BASE_IMAGE_OS ?= $(OSB_BASEOS)
OSB_IMAGE_PREFIX ?= crunchydata
OSB_ROOT ?= $(CURDIR)
PACKAGER ?= yum
OSB_VERSION ?= 4.7.1

PULL ?= IfNotPresent

# Determines whether or not images should be pushed to the local docker daemon when building with
# a tool other than docker (e.g. when building with buildah)
IMG_PUSH_TO_DOCKER_DAEMON ?= true

DOCKERBASEREGISTRY=registry.access.redhat.com/

ifeq ("$(OSB_BASEOS)", "ubi8")
        PACKAGER=microdnf
        BASE_IMAGE_OS=ubi8-minimal
endif

ifeq ("$(OSB_BASEOS)", "centos7")
        DOCKERBASEREGISTRY=centos:
endif

ifeq ("$(OSB_BASEOS)", "centos8")
        PACKAGER=dnf
        DOCKERBASEREGISTRY=centos:
endif

build: ## Builds the starter pack
	go build -i github.com/crunchydata/pgo-osb/cmd/pgo-osb

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o pgo-osb --ldflags="-s" github.com/crunchydata/pgo-osb/cmd/pgo-osb
main:
	go install pgo-osb.go

copy-bin:
	cp $(GOBIN)/pgo-osb .

buildah-image:
	sudo --preserve-env buildah bud --squash \
		-f $(OSB_ROOT)/build/pgo-osb/Dockerfile \
		-t $(OSB_IMAGE_PREFIX)/pgo-osb:$(OSB_IMAGE_TAG) \
		--build-arg BASEOS=$(OSB_BASEOS) \
		--build-arg DFSET=$(DFSET) \
		--build-arg DOCKERBASEREGISTRY=$(DOCKERBASEREGISTRY) \
		--build-arg BASE_IMAGE_OS=$(BASE_IMAGE_OS) \
		--build-arg PACKAGER=$(PACKAGER) \
		--build-arg RELVER=$(OSB_VERSION) \
		$(OSB_ROOT)


image: main license copy-bin buildah-image ;
# only push to docker daemon if variable PGO_PUSH_TO_DOCKER_DAEMON is set to "true"
ifeq ("$(IMG_PUSH_TO_DOCKER_DAEMON)", "true")
	sudo --preserve-env buildah push $(OSB_IMAGE_PREFIX)/pgo-osb:$(OSB_IMAGE_TAG) docker-daemon:$(OSB_IMAGE_PREFIX)/pgo-osb:$(OSB_IMAGE_TAG)
endif

push:
	docker push $(OSB_IMAGE_PREFIX)/pgo-osb:$(OSB_IMAGE_TAG)


deploy:
	cd deploy && ./deploy.sh

clean: ## Cleans up build artifacts
	rm -f pgo-osb
	rm -f pgo-osb-linux
	rm -f image/pgo-osb

provision: ## Provisions a service instance
	expenv -f manifests/service-instance.yaml | kubectl create -f -
deprovision:
	kubectl delete serviceinstance testinstance -n ${OSB_NAMESPACE}
provision2: ## Provisions a service instance
	expenv -f manifests/service-instance2.yaml | kubectl create -f -
deprovision2:
	kubectl delete serviceinstance testinstance2 -n ${OSB_NAMESPACE}

setup:
	go get github.com/blang/expenv

bind: ## Creates a binding
	expenv -f manifests/service-binding.yaml | kubectl create -f -
bind2: ## Creates a binding
	expenv -f manifests/service-binding2.yaml | kubectl create -f -

license: ## Aggregate all of the license files used to build the go binary
	./bin/license_aggregator.sh

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: deploy build test linux image clean push deploy-helm deploy-openshift create-ns provision bind license help
