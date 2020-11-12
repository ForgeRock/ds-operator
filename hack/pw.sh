#!/usr/bin/env bash
# Get the admin password
kubectl get secret ds-passwords -o jsonpath="{.data.dirmanager\\.pw}" | base64 --decode