 # DS Operator - experimental / POC

## Hacking

Note: Secrets are not generated yet - so run another deployment to get secrets generated (e.g. ds-only, all, etc)

```bash
# We are not using webhooks right now...
export ENABLE_WEBHOOKS="false"
# for running locally instead of in the cluster
export DEV_MODE=true
make install
make run
kubectl apply -f hack/ds.yaml
kubectl scale directoryservice/ds --replicas=2
kubectl delete -f hack/ds.yaml
```

You must provide the following secrets:

* cloud-storage-credentials - can be dummy values if you dont use cloud backup, but the secret is still required.
* ds - keystore with the master keystore and pin. Created by Secret Agent

## Design notes / philosophy

* Avoid just reimplementing the kustomize deployment as an operator. The operator should make some
  opiionated choices about how DS gets deployed. Ideally covering most use cases, but not attempting to cover all.
* Support bring your own secrets (secret agent) as well as operator generated secrets.


## What works now

* basic statefulset is created
* headless service created
* Deleting the CR properly cleans up the statefulset and service (owner refs are working OK). PVC is left behind - which is a good thing
* Scale subresource support (kubectl scale directoryservice/ds --replicas=2)
* Service account passwords are now suported. The operator can change the acount password.
* backup / restore implemented

Updating the spec.image will update the statefulset and perform a rolling update. For example:

```bash
kubectl patch directoryservice/ds --type='json' \
   -p='[{"op": "replace", "path": "/spec/image", "value":"gcr.io/forgeops-public/ds-idrepo:2020.10.28-AlSugoDiNoci"}]'
```

## What doesn't work

* No load balancer
* Limited status updated on CR object
* Few modifications when the CR changes. Need to understand what changes in the CR can be reflected in the sts:
  * spec.Replicas should increase the number of sts replicas
  * spec.Image should update the base image and do a rolling update
  * Changes to spec.LoadBalancer?

## Optional Features

* snapshots using k8s snapshots?
* Updates - patches
* Return dsrepl status for CR status updates
* DS proxy with affinity
* SSL / client auth

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


## DS JIRAs to track

* https://bugster.forgerock.org/jira/browse/OPENDJ-7582


## Implementation Notes


Spec update: https://kubernetes.slack.com/archives/CAR30FCJZ/p1602800878040500?thread_ts=1602647971.012900&cid=CAR30FCJZ
"One safe pattern is to mutate the spec, then update (i.e. commit) the spec, then mutate the status, then commit the status."

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




Sample bin/status output

disk free space - /opt/opendj/db

Backend info:

Base DN                       : Entries : Replication : Receive delay : Replay delay : Backend         : Type  : Active cache



cn=monitor

ds-mon-disk-root=/opt/opendj/db,cn=disk space monitor,cn=monitor
ds-mon-disk-free   - disk free space on /opt/opendj/db (data partition). Specific to the node queried, but should be relatively equal


cn=jvm,cn=monitor
ds-mon-jvm-memory-heap-used
ds-mon-jvm-memory-heap-max



ds-mon-domain-name=dc=openidm\,dc=forgerock\,dc=io,cn=replicas,cn=replication,cn=monitor
ds-mon-status
objectclass: ds-monitor-replica (structural)


ds-mon-server-id=ds-0,cn=servers,cn=topology,cn=monitor
objectclass: ds-monitor-topology-server (structural)
ds-mon-replication-domain - multi value- each entry is a dn that is replicated. ou=identities, etc.

