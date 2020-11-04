# ForgeRock Directory Service Operator - ds-operator

The ds-operator deploys and manages the ForgeRock Directory Server into a namespace on a Kubernetes cluster.

**This is an early alpha project and should not be used in production.**

Please see the annotated [hack/ds.yaml](hack/ds.yaml) for the up to date reference on
the format of the DirectoryService custom resource.

Developers please refer to the [developers guide](DEVELOPMENT.md)


## Deployment

The operator is deployed with the following command:

```bash
kubectl apply -f config/default
```

The operator runs in the `fr-system` namespace. To see log output use kubectl:

```bash
kubectl -n fr-system get pods
kubectl -n fr-system logs ds-operator-xxxx -f
# or if you have stern installed...
stern -n fr-system ds-
```

Only one instance of the operator is required per cluster.

## Operation

Once the ds-operator has been installed, you can deploy a directory service instance using:

```bash
# Sample ds deployment
kubectl apply -f hack/ds.yaml

# Scale the deployment:
kubectl scale directoryservice/ds --replicas=2
```

The directory service deployment creates a statefulset to run the directory service. The usual
`kubectl` commands (get, describe) can be used to diagnose the statefulset, pods, and services.

## Secrets

The operator supports creating (some) secrets, or a bring your own secrets model. Currently the operator CAN NOT generate the
keystore required for the directory service. Use [secret agent](https://github.com/ForgeRock/secret-agent) for that. If you have a ForgeOps deployment
the secrets created for the directory service are compatible with the operator.

## Cloud Storage Credentials

The operator can backup and restore from cloud storage on GCP, AWS and Azure. A secret must be
provided that contains the credentials required to backup or restore to the cloud.

The secret must contain one or more of the following key/value pairs:

```yaml
AZURE_ACCOUNT_KEY:       # Update if using Azure cloud storage for DS Backups
AZURE_ACCOUNT_NAME:      # Update if using Azure cloud storage for DS Backups
AWS_ACCESS_KEY_ID:       # Update if using AWS cloud storage for DS Backups
AWS_SECRET_ACCESS_KEY:   # Update if using AWS cloud storage for DS Backups
GOOGLE_CREDENTIALS:      #  google credentials.json format - Update if using GCP cloud storage for DS Backups
```

If you do not wish to use this feature, the operator will create a dummy  `cloud-storage-credentials` secret which
you can ignore. It will be deleted when the custom resource is deleted.

For GCP, a sample script [create-gcp-creds.sh](hack/create-gcp-creds.sh) is provided that will create a storage bucket and a
service account that can access that bucket. It also creates kubernetes  `cloud-storage-credentials` secret.

See the annotated resource for more information

## Scheduling

The operator provides the following control over scheduling:

* The pods will tolerate nodes that have been tainted with the following: `kubectl taint nodes node1 key=directory:NoSchedule`. Tainting
 such nodes will help them "repel" any non DS images, which can be helpful for performance.  If nodes are not tainted,
 this toleration has no effect. This should be thought of as an optional feature that most users will not require.
* Anti-Affinity: The DS nodes will prefer to be scheduled on nodes that do not have other directory pods on them. This is
  "soft" anti-affinity.  DS pods will be scheduled on the same node if the scheduler is not able to fulfill the request.