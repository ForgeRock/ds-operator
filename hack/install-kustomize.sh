#!/usr/bin/env bash
if type -P kustomize &> /dev/null;
then
	echo "kustomize installed, not installing";
	exit;
fi

case "${OSTYPE}" in
    "darwin"*) os="darwin";;
    "linux"*) os="linux";;
esac

latest=$(curl -s -w "%{redirect_url}" https://github.com/kubernetes-sigs/kustomize/releases/latest -o /dev/null | awk -F "/v" '{ printf $NF }')
if ! curl -o /tmp/kustomize.tar.gz -L "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${latest}/kustomize_v${latest}_${os}_amd64.tar.gz";
then
    echo "failed to download kustomize"
    exit 1;
fi
if ! tar vxf /tmp/kustomize.tar.gz -C /tmp/;
then
    echo "failed to extract kustomize"
    exit 1;
fi
if ! install /tmp/kustomize /go/bin/kustomize
then
    echo "failed to install to bin"
    exit 1;
fi
if ! rm -rf /tmp/{kustomize,kustomize.tar.gz};
then
    echo "failed to cleanup"
    exit 1;
fi
