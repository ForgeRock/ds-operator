#!/usr/bin/env bash

JSON=$(cat ~/etc/GOOGLE_CREDENTIALS_JSON)

kubectl delete secret cloud-storage-credentials
kubectl create secret generic cloud-storage-credentials --from-literal="gcp-credentials.json=$JSON"
