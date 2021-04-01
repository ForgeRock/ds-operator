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


## Volume Snapshots (Preview)

Beginning in Kubernetes 1.20, [Volume Snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) are generally
available. Snapshots enable the creation of a rapid point-in-time snapshot of a disk image. This feature
may also be available in earlier verions going back to 1.17.

The ds-operator enables the following snapshot features:

* The ability to initialize directory server data from a previous volume snapshot. This process
is much faster than recovering from backup media. The time to clone the disk will depend on the provider,
but in general will happen at block I/O speeds.
* The ability to schedule snapshots at regular intervals, and to automatically delete older snapshots.

Use cases that are enabled by snapshots include:

* Rapid rollback and recovery from the last snapshot point.
* For testing, initializing the directory with a large amount of sample data saved in a previous snapshot.
* In the future, snapshots can enable backups in a pod that is not serving traffic.

### Snapshot Prerequisites

* The cloud provider must support Kubernetes Volume Snapshots. This has been tested on GKE version 1.18.
* Snapshots require the `csi` volume driver. Consult your providers documentation. On GKE enable
  the `GcePersistentDiskCsiDriver` addon when creating or updating the cluster. The forgeops `cluster-up.sh` script
  for GKE has been updated to include this addon.
* Create a `VolumeSnapshotClass`. The default expected by the ds-operator is `ds-snapshot-class`. The `cluster-up.sh` script also creates
  this class using the following definition
```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshotClass
metadata:
  name: ds-snapshot-class
driver: pd.csi.storage.gke.io
deletionPolicy: Delete
EOF
```

* The StorageClass in the ds-operator deployment yaml must also use the CSI driver. When enabling the `GcePersistentDiskCsiDriver` addon, GKE will automatically
  create two new storage classes: `standard-rwo` (balanced PD Disk) and `premium-rwo` (SSD PD disk). The example `hack/ds.yaml`  has been updated.

### Initalizing the Directory server from a previous Volume Snapshot

Edit the Custom Resource Spec (CR), and under spec, add the following:

```
spec:
  initializeFromSnapshotName: "latest"
```

The field `spec.initializeFromSnapshotName` is optional. If present, it is the name of a VolumeSnapshot that will be used
to initialze the directory PVC claims. The special value "latest" will be interpreted by the operator as the
"latest" snapshot that the operator made (assuming automatic snapshots are enabled).

Notes:

* If a PVC claim is already provisioned, this setting has no effect. This setting only controls
initialization of *new* PVC claims. Said another way, existing data will not be overwritten by the snapshot.
* Changing this setting after deployment of a directory instance does not have any effect. This
setting works by updating the volume claim template for the Kubernetes StatefulSet. Kubernetes
does not allow the template to be updated after deployment.


A "rollback" procedure is as follows:

* Validate that you have a good snapshot to rollback to. `kubectl get volumesnapshots`.
* Delete the directory deployment *and* the PVC claims.  For example: `kubectl delete directoryservice/ds-idrepo-0 && kubectl delete pvc ds-idrepo-0` (repeat for all PVCs).
* Set the `spec.initializeFromSnapshotName` either to "latest" (assuming the operator is taking snapshots) or the name of a specific volume snapshot.
* Redeploy the directory service.  The new PVCs will be cloned from the snapshot.
* *All* directory instances are initialized from the same snapshot volume. For example, in a two way replication topology,
  ds-idrepo-0 and ds-idrepo-1 will both contain the same starting data and changelog.

*WARNING*: This procedure destroys any updates made since the last snapshot. Use with caution.

### Enabling Automatic VolumeSnapshots

The ds-operator can be configured to automatically take snapshots. Update the CR spec:

```yaml
snapshots:
  enabled: true
  # Take a snapshot every 20 minutes
  periodMinutes: 20
  # Keep this many snapshots. Older snapshots will be deleted
  snapshotsRetained: 3
  # This defaults to ds-snapshot-class if not specified
  volumeSnapshotClassName: ds-snapshot-class
```

Snapshot settings can be dynamically changed while the directory is running.

Notes:

* Only the first pvc (data-ds-idrepo-0 for example) is used for the snapshot. When initializing from a snapshot,
  all directory replicas start with the same data (see above)
* Snapshots can be expensive. Do not snapshot overly frequently, and retain only the number of
  snapshots that you need for availability.
* Snapshots are not a replacement for offline backups. If the data on disk is corrupt, the snapshot will also
  be corrupt.
* The very first snapshot will not happen until after the first `periodMinutes` (20 minutes in the example above).
  This is to give the directory time to start up before taking the first snapshot.

### Advanced Snapshot Scenarios

You can manually take snapshots at any time assuming you are using the CSI driver. For example

```
kubectl apply -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: my-great-snapshot-1
spec:
  volumeSnapshotClassName: ds-snapshot-class
  source:
    persistentVolumeClaimName: data-ds-idrepo-0
EOF
```

You can pre-create PVC claims, causing Kubernetes to ignore the PVC claim template created by the operator.
For example, if you wanted to initialize a 3rd directory instance with a specific snapshot, apply
the following *before* the PVC is created by the operator:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    pv.beta.kubernetes.io/gid: "0"
  labels:
  name: data-ds-idrepo-2
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: standard-rwo
  dataSource:
    apiGroup: snapshot.storage.k8s.io
    kind: VolumeSnapshot
    name: my-cool-snap-1
```
