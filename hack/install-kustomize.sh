#!/usr/bin/env bash
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
# Make kustomize available on the path
mv kustomize /bin && chmod a+rx /bin/kustomize
