# Needed for Travis - it won't like the version regex check otherwise
SHELL=/bin/bash

# Identifies the current build.
# These will be embedded in the app and displayed when it starts.
VERSION ?= 0.0.1-SNAPSHOT
COMMIT_HASH ?= $(shell git rev-parse HEAD)

# Indicates which version of the UI console is to be embedded
# in the docker image. If "local" the CONSOLE_LOCAL_DIR is
# where the UI project has been git cloned and has its
# content built in its build/ subdirectory.
# WARNING: If you have previously run the 'docker' target but
# later want to change the CONSOLE_VERSION then you must run
# the 'clean' target first before re-running the 'docker' target.
CONSOLE_VERSION ?= latest
CONSOLE_LOCAL_DIR ?= ../ui

# Version label is used in the OpenShift/K8S resources to identify
# their specific instances. These resources will have labels of
# "app: canary" and "version: ${VERSION_LABEL}" In this development
# environment, setting the label names equal to the branch names
# allows developers to build and deploy multiple Canary instances
# at the same time which is useful for debugging and testing.
# Due to restrictions on allowed characters in label values,
# we ensure only alphanumeric, underscore, dash, and dot are allowed.
# Due to restrictions on allowed characters in names, we convert
# uppercase characters to lowercase characters and
# underscores/dots/spaces to dashes.
# If we are deploying from a branch, we must ensure all the OS/k8s
# names to be created are unique - the NAME_SUFFIX will be appended
# to all names so they are unique.
VERSION_LABEL ?= $(shell git rev-parse --abbrev-ref HEAD)
ifneq ($(shell [[ ${VERSION_LABEL} =~ ^[a-zA-Z0-9]([-_.a-zA-Z0-9]*[a-zA-Z0-9])?$$ ]] && echo valid),valid)
  $(error Your version label value '${VERSION_LABEL}' is invalid and cannot be used.)
endif
ifeq (${VERSION_LABEL},master)
  NAME_SUFFIX ?=
else
  # note: we want to start suffix with a dash so it is between the name and the suffix
  NAME_SUFFIX ?= $(shell echo -n "-${VERSION_LABEL}" | tr '[:upper:][:space:]_.' '[:lower:]---')
  ifneq ($(shell [[ ${NAME_SUFFIX} =~ ^([-a-z0-9]*[a-z0-9])?$$ ]] && echo valid),valid)
    $(error Your name suffix '${NAME_SUFFIX}' is invalid and cannot be used.)
  endif
endif

# The minimum Go version that must be used to build the app.
GO_VERSION_CANARY = 1.8.3

# Identifies the docker image that will be built and deployed.
# Note that if building from a non-master branch, the default
# version will be the same name as the branch allowing you to
# deploy different builds at the same time.
DOCKER_NAME ?= devtio/canary
ifeq  ("${VERSION_LABEL}","master")
  DOCKER_VERSION ?= dev
else
  DOCKER_VERSION ?= ${VERSION_LABEL}
endif
DOCKER_TAG = ${DOCKER_NAME}:${DOCKER_VERSION}

# Indicates the log level the app will use when started.
# <4=INFO
#  4=DEBUG
#  5=TRACE
VERBOSE_MODE ?= 4

# Declares the namespace where the objects are to be deployed.
# For OpenShift, this is the name of the project.
NAMESPACE ?= canary

# Environment variables set when running the Go compiler.
GO_BUILD_ENVVARS = \
	GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=0 \

clean:
	@echo Cleaning...
	@rm -rf ${GOPATH}/bin/devtio/canary
	@rm -rf ${GOPATH}/pkg/*
	@rm -rf _output/docker

build:
	@echo Building...
	${GO_BUILD_ENVVARS} go build \
		-o ${GOPATH}/bin/devtio/canary

.prepare-docker-image-files: 
	@echo Preparing docker image files...
	@mkdir -p _output/docker
	@cp -r deploy/docker/* _output/docker
	@cp ${GOPATH}/bin/devtio/canary _output/docker

.prepare-minikube:
	@minikube addons list | grep -q "ingress: enabled" ; \
	if [ "$$?" != "0" ]; then \
		echo "Enabling ingress support to minikube" ; \
		minikube addons enable ingress ; \
	fi
	@grep -q canary /etc/hosts ; \
	if [ "$$?" != "0" ]; then \
		echo "/etc/hosts should have canary so you can access the ingress"; \
	fi

minikube-docker: .prepare-minikube .prepare-docker-image-files
	@echo Building docker image into minikube docker daemon...
	@eval $$(minikube docker-env) ; \
	docker build -t ${DOCKER_TAG} _output/docker

gcr-docker: .prepare-docker-image-files
	@echo Building docker image into minikube docker daemon...
	docker build -t ${DOCKER_TAG} _output/docker
	docker tag ${DOCKER_TAG} ${GCR_ID}
	docker push ${GCR_ID}

k8s-deploy: k8s-undeploy
	@if ! which envsubst > /dev/null 2>&1; then echo "You are missing 'envsubst'. Please install it and retry. If on MacOS, you can get this by installing the gettext package"; exit 1; fi
	@echo Deploying to Kubernetes namespace ${NAMESPACE}
	cat deploy/kubernetes/canary-configmap.yaml | VERSION_LABEL=${VERSION_LABEL} NAME_SUFFIX=${NAME_SUFFIX} envsubst | kubectl create -n ${NAMESPACE} -f -
	kubectl label namespace ${NAMESPACE} istio-injection=enabled --overwrite	
	cat deploy/kubernetes/canary.yaml | GCR_ID=${GCR_ID} GATEWAY_URL=${GATEWAY_URL} IMAGE_NAME=${GCR_ID} IMAGE_VERSION=latest NAMESPACE=${NAMESPACE} VERSION_LABEL=${VERSION_LABEL} NAME_SUFFIX=${NAME_SUFFIX} VERBOSE_MODE=${VERBOSE_MODE} envsubst | kubectl create -n ${NAMESPACE} -f -

k8s-undeploy:
	@echo Undeploying from Kubernetes namespace ${NAMESPACE}
	kubectl delete all,secrets,sa,configmaps,deployments,ingresses,clusterroles,clusterrolebindings,routerules,virtualservices,gateways --selector=app=canary --selector=version=${VERSION_LABEL} -n ${NAMESPACE}


#
# dep targets - dependency management
#

dep-install:
	@echo Installing Glide itself
	@mkdir -p ${GOPATH}/bin
	# We want to pin on a specific version
	# @curl https://glide.sh/get | sh
	@curl https://glide.sh/get | awk '{gsub("get TAG https://glide.sh/version", "TAG=v0.13.1", $$0); print}' | sh

dep-update:
	@echo Updating dependencies and storing in vendor directory
	@glide update --strip-vendor
