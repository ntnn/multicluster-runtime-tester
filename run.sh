#!/usr/bin/env sh

set -e

make manifests generate fmt vet

export KUBECONFIG="$PWD/kubeconfig.yaml"
if ! test -f "$KUBECONFIG"; then
    kind create cluster --kubeconfig "$KUBECONFIG"
fi

make install

make run
