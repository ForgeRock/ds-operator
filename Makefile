.PHONY: build install run test uninstall deploy manifest fmt vet generate docker-build release
# Image URL to use all building/pushing image targets
DEFAULT_IMG = gcr.io/engineering-devops/ds-operator
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
#CRD_OPTIONS ?= "crd:trivialVersions=false"
# This will work on kube versions 1.16+. We want the CRD OpenAPI validation features in v1
CRD_OPTIONS ?= "crd:crdVersions=v1"

# If IMG not set and PR not set, set it to latest
ifdef PR_NUMBER
IMG ?= "${DEFAULT_IMG}:pr-${PR_NUMBER}"
endif
IMG ?= "${DEFAULT_IMG}:latest"
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
# Use --zap-log-level 10 to set detailed trace
run: generate fmt vet manifests
	go run ./main.go --zap-devel

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

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

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

# Test and Build container
docker-build: test build
	@echo "${IMG} built"

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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Integration tests assume and ldap server is running via a localhost:1389 proxy
int:
	cd pkg/ldap && go test --tags=integration
