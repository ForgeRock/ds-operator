# ForgeRock Directory Service Operator - ds-operator

The ds-operator deploys
the [ForgeRock Directory Server](https://www.forgerock.com/platform/directory-services)
 in a Kubernetes cluster. This
is an implementation of the [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) pattern.

Basic features of the operator include:

* Creation of StatefulSets, Services and Persistent volume claims for the directory
* Configures replication by adding new directory pods to the replication topology
* Backup and restore of the directory to cloud storage such as AWS S3, GCP bukets or Azure storage
* Change service account passwords in the directory using a Kubernetes secret.


**Note: This is an early alpha project and should not be used in production.**

Please see the annotated [hack/ds.yaml](hack/ds.yaml) for the most current reference on the DirectoryService custom resource specification.

Developers please refer to the [developers guide](DEVELOPMENT.md)


## Install the Operator

**Important: Only one instance of the operator is required per cluster.**

ForgeRock developers: This is already installed on the `eng-shared` cluster.

The [install.sh](install.sh) script will install the latest release of the operator. You can also curl this script:

```bash
curl  -L  "https://github.com/ForgeRock/ds-operator/releases/latest/download/install.sh" -o /tmp/install.sh
chmod +x /tmp/install.sh
/tmp/install.sh install
```

The operator runs in the `fr-system` namespace. To see log output use kubectl:

```bash
kubectl -n fr-system get pods
kubectl logs -n fr-system  -l  control-plane=ds-operator -f
# or if you have stern installed...
stern -n fr-system ds-
```

## Install the Secret Agent Operator

The ds-operator can create some (but not all) secrets needed by the directory server.
Use the [secret agent operator](https://github.com/ForgeRock/secret-agent) to
create the required secrets.

Secret agent can be installed using the [secret-agent.sh](https://raw.githubusercontent.com/ForgeRock/forgeops/master/bin/secret-agent.sh) script.

* Note to ForgeRock developers: secret-agent is already installed on the eng-shared cluster.*

## Deploy a Directory Instance

Once the ds-operator has been installed, you can deploy an instance of the directory service using the custom
resource provided in `hack/ds.yaml`.

Below is a sample deployment session

```bash
# Create the required secrets using secret agent. If you get an error here check
# to see that you have deployed secret agent.
kubectl apply -f hack/secret_agent.yaml
kubectl get secrets

# Deploy the sample directory instance
kubectl apply -f hack/ds.yaml

# pw.sh script will retrieve the uid=admin password:
./hack/pw.sh

# View the pods, statefulset, etc
kubectl get pod

# Scale the deployment by adding another replica
kubectl scale directoryservice/ds-idrepo --replicas=2

# You can edit the resource, or edit the ds.yaml and kubectl apply changes
# Things you can change at runtime include the number of replicas, enable/disable of backup/restore
kubectl edit directoryservice/ds-idrepo

# Delete the directory instance.
kubectl delete -f hack/ds.yaml

# If you want to delete the PVC claims...
kubectl delete pvc data-ds-0
```

The directory service deployment creates a statefulset to run the directory service. The usual
`kubectl` commands (get, describe) can be used to diagnose the statefulsets, pods, and services.

## Directory Docker Image

The deployed spec.image must support the `operator-init.sh` entrypoint. There is a sample skaffold file in `forgeops` that will build
and push an image:

```bash
cd forgeops/docker/7.0/ds
skaffold --default-repo gcr.io/engineering-devops build
```

Evaluation images have been built for you on gcr.io/forgeops-public/ds-idrepo. The [ds.yaml](hack/ds.yaml) Custom Resource references this image.

The operator assume the referenced image is built for purpose, and has all the required configuration, indexes and schema for your deployment.

## Secrets

The operator supports creating (some) secrets, or a bring-your-own secrets model. Currently the operator *CAN NOT* generate the
keystore required for the directory service. Use [secret agent](https://github.com/ForgeRock/secret-agent) for that. If you have a ForgeOps deployment
the secrets created for the existing directory service are compatible with the operator.

The operator can generate random secrets for the `uid=admin` account, `cn=monitor` and application service accounts (for
example - the AM CTS account). Refer to the annotated sample.

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
service account that can access that bucket. It also creates the Kubernetes  `cloud-storage-credentials` secret.

Note to ForgeRock developers: A GCP bucket has been created for you. Reach out on slack.

See the annotated custom resource for more information

## Scheduling

The operator provides the following control over scheduling:

* The pods will tolerate nodes that have been tainted with the following:
  * `kubectl taint nodes node1 key=directory:NoSchedule`.
* Tainting
 such nodes will help them "repel" non directory workloads, which can be helpful for performance.  If nodes are not tainted,
 this toleration has no effect. This should be thought of as an optional feature that most users will not require.
* Anti-Affinity: The DS nodes will prefer to be scheduled on nodes that do not have other directory pods on them. This is
  "soft" anti-affinity.  DS pods will be scheduled on the same node if the scheduler is not able to fulfill the request.