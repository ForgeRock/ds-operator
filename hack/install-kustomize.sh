#!/usr/bin/env bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
# Make kustomize available on the path
cp kustomize /bin && chmod a+rx /bin/kustomize
