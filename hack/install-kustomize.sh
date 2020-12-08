#!/usr/bin/env bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
# Make kustomize available on the gopath/bin - where the make process can find it
cp kustomize $(go env GOPATH)/bin

