#!/usr/bin/env bash
#
# Copyright 2019-2020 ForgeRock AS. All Rights Reserved
#
# Use of this code requires a commercial software license with ForgeRock AS.
# or with one of its affiliates. All use shall be exclusively subject
# to such license between the licensee and ForgeRock AS.

set -eu

# ParallelGC with a single generation tenuring threshold has been shown to give the best
# performance vs determinism trade-off for servers using JVM heaps of less than 8GB,
# as well as all batch tool use-cases such as import-ldif.
# Unusual deployments, such as those requiring very large JVM heaps, should tune this setting
# and use a different garbage collector, such as G1.
# The /dev/urandom device is up to 4 times faster for crypto operations in some VM environments
# where the Linux kernel runs low on entropy. This settting does not negatively impact random number security
# and is recommended as the default.
DEFAULT_OPENDJ_JAVA_ARGS="-XX:MaxRAMPercentage=75 -XX:+UseParallelGC -XX:MaxTenuringThreshold=1 -Djava.security.egd=file:/dev/urandom"
export OPENDJ_JAVA_ARGS=${OPENDJ_JAVA_ARGS:-${DEFAULT_OPENDJ_JAVA_ARGS}}

export DS_GROUP_ID=${DS_GROUP_ID:-default}
export DS_SERVER_ID=${DS_SERVER_ID:-${HOSTNAME:-localhost}}
export DS_ADVERTISED_LISTEN_ADDRESS=${DS_ADVERTISED_LISTEN_ADDRESS:-$(hostname -f)}

# If the advertised listen address looks like a Kubernetes pod host name of the form
# <statefulset-name>-<ordinal>.<domain-name> then derived the default bootstrap servers names as
# <statefulset-name>-0.<domain-name>,<statefulset-name>-1.<domain-name>.
#
# Sample hostnames from Kubernetes include:
#
#     ds-1.userstore.svc.cluster.local
#     ds-userstore-1.userstore.svc.cluster.local
#     userstore-1.userstore.jnkns-pndj-bld-pr-4958-1.svc.cluster.local
#     ds-userstore-1.userstore.jnkns-pndj-bld-pr-4958-1.svc.cluster.local
#
if [[ "${DS_ADVERTISED_LISTEN_ADDRESS}" =~ [^.]+-[0-9]+\..+ ]]; then
    podDomain=${DS_ADVERTISED_LISTEN_ADDRESS#*.}
    podName=${DS_ADVERTISED_LISTEN_ADDRESS%%.*}
    podPrefix=${podName%-*}

    ds0=${podPrefix}-0.${podDomain}:8989
    ds1=${podPrefix}-1.${podDomain}:8989
    export DS_BOOTSTRAP_REPLICATION_SERVERS=${DS_BOOTSTRAP_REPLICATION_SERVERS:-${ds0},${ds1}}
else
    export DS_BOOTSTRAP_REPLICATION_SERVERS=${DS_BOOTSTRAP_REPLICATION_SERVERS:-${DS_ADVERTISED_LISTEN_ADDRESS}:8989}
fi


validateImage() {
    # FIXME: fail-fast if database encryption has been used when the image was built (OPENDJ-6598).
    diff -q template/db/adminRoot/admin-backend.ldif db/adminRoot/admin-backend.ldif > /dev/null || {
        echo "The server cannot start because it appears that database encryption"
        echo "was enabled for a backend when the Docker image was built."
        echo "This feature is not yet supported when using Docker."
        exit 1
    }
}

# Initialize persisted data in the "data" directory if it is empty, using data from the data directories
# contained in the Docker image. The data directory contains the server's persisted state, including db directories,
# changelog, and locks. In dev environments it is expected to be an empty tmpfs volume whose content is lost after
# restart.
bootstrapDataFromImageIfNeeded() {
    for d in ${dataDirs}; do
        if [ -z "$(ls -A data/$d 2>/dev/null)" ]; then
            echo "Initializing \"data/$d\" from Docker image"
            mkdir -p data
            mv $d data
        fi
    done
}

linkDataDirectories() {
    # List of directories which are expected to be found in the data directory.
    dataDirs="db changelogDb locks var config"
    mkdir -p data
    ls -l data
    for d in ${dataDirs}; do
        if [[ ! -d "data/$d" ]]; then
            echo "initializing data/$d with the contents of the docker image"
            mv $d data
        else
            # the data/$d exists -we want to make sure it is used - not the one in the image
            # rename the docker directory so the link works.
            mv $d $d.docker
        fi
        echo "Linking $d to data/$d"
        ln -s data/$d
    done
}

# If the pod was terminated abnormally then lock file may not have been cleaned up.
removeLocks() {
    rm -f /opt/opendj/locks/server.lock
}

# Make it easier to run tools interactively by exec'ing into the running container.
setOnlineToolProperties() {
    mkdir -p ~/.opendj
    cp config/tools.properties ~/.opendj
}

upgradeDataAndRebuildDegradedIndexes() {

    # Build an array containing the list of pluggable backend base DNs by redirecting the command output to
    # mapfile using process substitution.
    mapfile -t BASE_DNS < <(./bin/ldifsearch -b cn=backends,cn=config -s one data/config/config.ldif "(&(objectclass=ds-cfg-pluggable-backend)(ds-cfg-enabled=true))" ds-cfg-base-dn | grep "^ds-cfg-base-dn" | cut -c17-)

    # Upgrade is idempotent, so it should have no effect if there is nothing to do.
    # Fail-fast if the config needs upgrading because it should have been done when the image was built.
    echo "Upgrading configuration and data..."
     ./upgrade --dataOnly --acceptLicense --force --ignoreErrors --no-prompt

    # Rebuild any corrupt/missing indexes.
    for baseDn in "${BASE_DNS[@]}"; do
        echo "Rebuilding degraded indexes for base DN \"${baseDn}\"..."
        rebuild-index --offline --noPropertiesFile --rebuildDegraded --baseDn "${baseDn}" > /dev/null
    done
}

preExec() {
    echo
    echo "Server configured with:"
    echo "    Group ID                        : $DS_GROUP_ID"
    echo "    Server ID                       : $DS_SERVER_ID"
    echo "    Advertised listen address       : $DS_ADVERTISED_LISTEN_ADDRESS"
    echo "    Bootstrap replication server(s) : $DS_BOOTSTRAP_REPLICATION_SERVERS"
    echo
}

waitUntilSigTerm() {
    trap 'echo "Caught SIGTERM"' SIGTERM
    while :
    do
       sleep infinity &
       wait $!
    done
}

init() {
    echo "initializing..."
    linkDataDirectories
    removeLocks
    upgradeDataAndRebuildDegradedIndexes
}

CMD="${1:-help}"
case "$CMD" in

# Used by ds-operator. Relocates all DS data to data/ directory.
# This is a fully mutable installation.
init)
    [[ -d data/db ]] && {
        echo "data/ directory contains data. setup skipped";
        # Init still needs to check the indexes.
        init
        exit 0;
    }
    linkDataDirectories

    # If the user supplied a setup script, run it.
    # Note - on K8S this is a symlink
    if [[ -L scripts/setup ]]; then
        echo "Executing user supplied setup"
        /opt/opendj/scripts/setup
        exit 0
    fi
    echo "Executing default-scripts/setup"
    /opt/opendj/default-scripts/setup
    ;;

# Special start for ds operator. Just needs to link the data directories
start)
    init
    preExec
    exec start-ds --nodetach
    ;;

dev)
    # Sleep until Kubernetes terminates the pod using a SIGTERM.
    echo "Connect using 'kubectl exec -it POD -- /bin/bash'"
    waitUntilSigTerm
    ;;

*)
    validateImage
    linkDataDirectories
    removeLocks
    preExec
    shift
    exec "$@"
    ;;

esac
