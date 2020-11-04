#!/usr/bin/env bash
# Sample script to create cloud credentials for GCP backup/restore


SVC="ds-operator"
PROJECT=$(gcloud config get-value project)
BUCKET="gs://ds-operator-${PROJECT}"

SVC_ACCOUNT="${SVC}@${PROJECT}.iam.gserviceaccount.com"

# Create the service account for ds-operator
gcloud iam service-accounts create "$SVC" --description="ds-operator SA for backup/restore to GCS" --display-name="ds-operator"

# create the bucket

gsutil mb "$BUCKET"

# Enable the service account admin access to the bucket
gsutil iam ch "serviceAccount:${SVC_ACCOUNT}:objectAdmin" "$BUCKET"

# Create a key for the service account
gcloud iam service-accounts keys create /var/tmp/sa.json --iam-account "$SVC_ACCOUNT"

JSON=$(cat /var/tmp/sa.json)

echo "Creating Kubernetes secret"
kubectl delete secret cloud-storage-credentials

kubectl create secret generic cloud-storage-credentials --from-literal="gcp-credentials.json=$JSON"

echo "You should delete the service account key in /var/tmp/sa.json"