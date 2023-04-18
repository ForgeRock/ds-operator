#!/usr/bin/env bash
# Install the operator from a released version. Use the hack/install-operator script for adhoc testing.
#

# As releases are tagged, update the operator version here. Avoid using "latest" to ensure consistent behavior.
DS_OPERATOR_VERSION=${DS_OPERATOR_VERSION:-v0.2.6}

USAGE="Usage: $0 install|remove|upgrade"


URL="https://github.com/ForgeRock/ds-operator/releases/download/${DS_OPERATOR_VERSION}/ds-operator.yaml"

if [ "$DS_OPERATOR_VERSION" == "latest" ]; then
    URL="https://github.com/ForgeRock/ds-operator/releases/latest/download/ds-operator.yaml"
fi

install() {
    printf "Checking ds-operator and related CRDs: "
    if ! $(kubectl get crd directoryservices.directory.forgerock.io &> /dev/null); then
        printf "ds-operator not found. Installing ds-operator version: '${DS_OPERATOR_VERSION}'\n"
         kubectl apply -f "$URL"
    else
        printf "ds-operator CRD found in cluster. Skipping ds-operator installation.\n"
    fi
}



# Upgrade - same as install, but omit check for exiting installation
upgrade() {
    printf "Applying upgrade to ds-operator"
    kubectl apply -f "$URL"
}

remove() {
    echo "Warning this is destructive and will remove all managed ds instances"
    echo "Waiting 10 seconds before removing."
    sleep 10
    kubectl delete -f "$URL"
}

cmd=${1}

echo "Version: $DS_OPERATOR_VERSION"

case "${cmd}" in
    install) install;;
    remove) remove;;
    upgrade) upgrade;;
    *) echo "Error: Incorrect usage"
       echo $USAGE
       exit 1;;
esac