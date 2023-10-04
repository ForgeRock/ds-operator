.PHONY: build install run test uninstall deploy manifest fmt vet generate docker-build release
ifndef DEFAULT_BUILDX_BUILDER
DEFAULT_BUILDX_BUILDER=default
endif
# Image URL to use all building/pushing image targets
DEFAULT_IMG_REGISTRY = us-docker.pkg.dev
DEFAULT_IMG_REPOSITORY = forgeops-public/images
DEFAULT_IMG = ${DEFAULT_IMG_REGISTRY}/${DEFAULT_IMG_REPOSITORY}/ds-operator
# This will work on kube versions 1.16+. We want the CRD OpenAPI validation features in v1
CRD_OPTIONS ?= "crd:crdVersions=v1"

ifndef DEFAULT_IMG_TAG
ifdef PR_NUMBER
DEFAULT_IMG_TAG=pr-${PR_NUMBER}
else
DEFAULT_IMG_TAG=latest
endif
endif
IMG ?= "${DEFAULT_IMG}:${DEFAULT_IMG_TAG}"
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Run tests
test tests: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
# Use --zap-log-level 10 to set detailed trace
run: generate fmt vet manifests
	DEBUG_CONTAINER=true DEV_MODE=true go run ./main.go --zap-devel

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Deploys controller -enabling debug side cars. Use when running with the csi hosthpath provisioner.
# This runs an init container to chown the pvc volume to the forgerock user.
deploy-debug: manifests
	cd config/debug && kustomize edit set image controller=${IMG}
	kustomize build config/debug | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifest manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=config/crd/bases
	# $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases


# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# just build binary to dist/ AND container
build:
	IMG=${IMG} goreleaser --snapshot --rm-dist

docker: test
	docker build -t ${IMG} .

# Test and Build container
docker-build: build
 	@echo "${IMG} built"

docker-buildx-bake:
	REGISTRY=${DEFAULT_IMG_REGISTRY} REPOSITORY=${DEFAULT_IMG_REPOSITORY} BUILD_TAG=${DEFAULT_IMG_TAG} docker buildx bake --file=docker-bake.hcl --builder=${DEFAULT_BUILDX_BUILDER}

# Build, push, and create GitHub release
release: install-tools
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > ds-operator.yaml
	git checkout config # goreleaser requires a clean workspace
	IMG=${IMG} goreleaser


# Install tools
install-tools:
	./hack/install-goreleaser.sh
	./hack/install-kustomize.sh

# Push the docker image
push: docker-build
	docker push ${IMG}


CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0)



# go-get-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef


# Integration tests assume and ldap server is running via a localhost:1389 proxy
inttest:
	cd pkg/ldap && go test --tags=integration
