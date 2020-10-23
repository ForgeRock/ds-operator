 # DS Operator - experimental / POC

## Hacking

Note: Secrets are not generated yet - so run another deployment to get secrets generated (e.g. ds-only, all, etc)

```bash
# We are not using webhooks just yet...
export ENABLE_WEBHOOKS="false"
make install
make run
kubectl apply -f hack/ds.yaml
kubectl scale directoryservice/ds --replicas=2
kubectl delete -f hack/ds.yaml
```

## Design notes / philosophy

* We want to avoid just reimplementing the kustomize deployment as an operator. The operator should make some
  opiionated choices about how DS gets deployed. Ideally covering most use cases, but not attempting to cover all.
* Support bring your own secrets (secret agent) as well as operator generated secrets.


## What works now

* basic statefulset is created and comes up OK
* headless service created
* Deleting the CR properly cleans up the statefulset and service (owner refs are working OK). PVC is left behind - which is a good thing
* Scale subresource support (kubectl scale directoryservice/ds --replicas=2)
* Service account passwords are now suported. The operator can change the acount password.

## What doesn't work

* No webhooks or validation
* No backup/restore yet
* No load balancer
* Limited status updated on CR object
* Few modifications when the CR changes. Need to understand what changes in the CR can be reflected in the sts:
  * spec.Replicas should increase the number of sts replicas
  * spec.Image should update the base image and do a rolling update
  * Changes to spec.LoadBalancer?

## Optional Features

* snapshots using k8s snapshots?
* Updates

See https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/#updating-statefulsets
In 1.17, some sts settings can be updated: image, Resource req/limit, labels and annotations. Might be useful to adjust JVM memory
* Alerts?
* Disk full alerts?
* Tuning. How do we tune backend params, or is that a function of the docker image

* cli tool that can run commands in ds. For example, running dsconfig / dsbackup commands.  cli could grab the admin creds to make this simpler.
**
* Snapshots - when K8S snapshots are widely supported, should we use snapshots for backup?

## Life-cycle hooks

* recover - initialize only if no backend data exists
  * backup - init from backup
  * ldif - init from an ldif file supplied as a configmap
* initialize - restore / repair any indexes
* run - running as a service
* maintenance - not servicing, but available for maintenance. Not sure if this is needed, or how to implement
* shutdown

## Scheduling

The operator provides the following control over scheduling:

* The pods will tolerate nodes that have been tainted like the following: `kubectl taint nodes node1 key=directory:NoSchedule`. Tainting
 such nodes will help them "repel" any non DS images, which can be helpful for performance.  If nodes are not tainted,
 this toleration has no effect. This should be thought of as an optional feature that most users will not require.
* Anti-Affinity: The DS nodes will prefer to be scheduled on nodes that do not have other directory pods on them. This is
  "soft" anti-affinity.  DS pods will be scheduled on the same node if the scheduler is not able to fulfill the request.


## Implementation Notes


Spec update: https://kubernetes.slack.com/archives/CAR30FCJZ/p1602800878040500?thread_ts=1602647971.012900&cid=CAR30FCJZ


If I parse that, it means you should do Status().Update() first (to not lose your status), and then Update()?

negz  16 hours ago
@warren.strange Kind of - the inverse is also true though. If you call Status().Update() you’ll lose any uncommitted changes to the spec or metadata.


negz  16 hours ago
So it’s less of a hard and fast ordering rule vs something to be aware of.

One safe pattern is to mutate the spec, then update (i.e. commit) the spec, then mutate the status, then commit the status.

dsbackup  \
--storageProperty gs.credentials.env.var:GOOGLE_CREDENTIALS \
 --backupLocation gs://forgeops/dj-backup/wstest


 dsbackup create \
 --hostname localhost \
 --port 4444 \
 --bindDN uid=admin \
 --bindPassword "xetvjwgos5e75pty0e5w3vnbpk3nwt1e" \
-X \
--storageProperty gs.credentials.path:/var/tmp/sa.json \
 --recurringTask "*/5 * * * *" \
 --taskId NightlyBackup \
--backupLocation gs://forgeops/dj-backup/wstest


# This is really slow...
dsbackup list \
--noPropertiesFile \
--storageProperty gs.credentials.path:/var/run/secrets/cloud-credentials-cache/GOOGLE_CREDENTIALS \
--backupLocation gs://forgeops/dj-backup/wstest \
--verify --last


 dsbackup restore \
 --hostname localhost \
 --port 4444 \
 --bindDN uid=admin \
 --bindPassword "xetvjwgos5e75pty0e5w3vnbpk3nwt1e" \
-X \
--storageProperty gs.credentials.path:/var/run/secrets/cloud-credentials-cache/GOOGLE_CREDENTIALS \
 --taskId NightlyRestore \
--backupLocation gs://forgeops/dj-backup/wstest \
--backendName amIdentityStore



Parsing DS zulu time: 20201021162043Z


https://backstage.forgerock.com/docs/ds/7/maintenance-guide/backup-restore.html#cloud-storage


Q: What happens if Istio injects a sidecar into the DS container? How does the operator make an ldap connection over mTLS


ds-task (structural)
ds-task-restore (structural)
top (abstract)
gs://forgeops/dj-backup/wstest
org.opends.server.tasks.RestoreTask
NightlyRestore
20201021224819Z
amIdentityStore
gs.credentials.path:/var/run/secrets/cloud-credentials-cache/GOOGLE_CREDENTIALS
[21/Oct/2020:22:48:19 +0000] category=BACKEND severity=NOTICE seq=0 msgID=413 msg=Restore task NightlyRestore started execution
RUNNING

