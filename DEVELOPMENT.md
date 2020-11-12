# Development notes - Directory Services Operator


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

When testing out of cluster, the controller on your desktop needs to open an ldap connection to the directory.
The variables DEV_MODE (see above) configures the controller connect to  localhost:1389.  In dev mode, port forward to the ds container:

```bash
kubectl port-forward ds-0 1389
```

You must provide the following secrets:

* ds - keystore with the master keystore and pin. Created by Secret Agent

## Design notes / philosophy

* Avoid just reimplementing the kustomize deployment as an operator. The operator should make some
  opionated choices about how DS gets deployed. Ideally covering most use cases, but not attempting to cover all.
* Support bring your own secrets (secret agent) as well as operator generated secrets.
* automate backup and restore, and eventually other administrative actions


## What works now

* basic statefulset is created
* headless service created
* Deleting the CR properly cleans up the statefulset and service (owner refs are working OK). PVC is left behind - which is a good thing
* Scale subresource support (`kubectl scale directoryservice/ds --replicas=2`)
* Service account passwords are now suported. The operator can change the account passwords for AM, IDM, etc..
* backup / restore implemented

Updating the spec.image will update the statefulset and perform a rolling update. For example:

```bash
kubectl patch directoryservice/ds --type='json' \
   -p='[{"op": "replace", "path": "/spec/image", "value":"gcr.io/forgeops-public/ds-idrepo:2020.10.28-AlSugoDiNoci"}]'
```

## What doesn't work

* No load balancer or ds proxy
* Limited status updated on CR object
* SSL certificate management
** Return dsrepl status for CR status updates



## Optional Features (Future...)

* Snapshots using k8s snapshots when supported on all clouds (Kubernetes 1.18)
* Updates - patches
* DS proxy with affinity
* SSL / client authentication
See https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/#updating-statefulsets
In 1.17, some sts settings can be updated: image, Resource req/limit, labels and annotations. Might be useful to adjust JVM memory
* Alerts?
* Disk full alerts?
* Tuning. How do we tune backend params, or is that a function of the docker image?
* cli tool that can run commands in ds. For example, running dsconfig / dsbackup commands.  cli can grab the admin creds to make this simpler.

## DS JIRAs to track

* https://bugster.forgerock.org/jira/browse/OPENDJ-7582  This will allow dynamic enable/disable of backups
* https://bugster.forgerock.org/jira/browse/OPENDJ-7502
* https://bugster.forgerock.org/jira/browse/OPENDJ-7501
* https://bugster.forgerock.org/jira/browse/OPENDJ-7352
* https://bugster.forgerock.org/jira/browse/CLOUD-2666
* 




## Implementation Notes

(Scratch Notes to implementers...)


Spec update: https://kubernetes.slack.com/archives/CAR30FCJZ/p1602800878040500?thread_ts=1602647971.012900&cid=CAR30FCJZ
"One safe pattern is to mutate the spec, then update (i.e. commit) the spec, then mutate the status, then commit the status."


Backup / restore notes

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


dsbackup purge  \
--hostname localhost \
 --port 4444 \
 --bindDN uid=admin \
 --bindPassword "xetvjwgos5e75pty0e5w3vnbpk3nwt1e" \
-X \
    --storageProperty gs.credentials.path:/var/run/secrets/cloud-credentials-cache/gcp-credentials.json \
    --backupLocation gs://ds-operator-engineering-devops/ds-backup-test \
     --taskId PurgeTask \
     --olderThan '12h' \
     --recurringTask "*/5 * * * *"


https://backstage.forgerock.com/docs/ds/7/maintenance-guide/backup-restore.html#cloud-storage


cn=monitor - things we might want to use for the status object:

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

