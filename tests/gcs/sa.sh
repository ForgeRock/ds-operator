
#!/usr/bin/env bash
# Create the service account for  gcs backup/restore jobs
GSA_NAME=ds-backup
K8S_NAMESPACE=default
KSA_NAME=ds-backup

kubectl create serviceaccount $KSA_NAME -n $K8S_NAMESPACE


kubectl annotate serviceaccount --namespace=$K8S_NAMESPACE  $KSA_NAME iam.gke.io/gcp-service-account=$GSA_NAME@engineering-devops.iam.gserviceaccount.com
PROJECT_ID=engineering-devops

gcloud iam service-accounts add-iam-policy-binding "$GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:$PROJECT_ID.svc.id.goog[$K8S_NAMESPACE/$KSA_NAME]"
