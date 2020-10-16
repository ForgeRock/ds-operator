 # DS Operator - experimental

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
 this toleration has no effect.
* Anti-Affinity: The DS nodes will prefer to be scheduled on nodes that do not have other directory pods on them. This is
  "soft" anti-affinity.  DS pods will be scheduled on the same node if the scheduler is not able to fulfill the request.



Backup:

./ldapmodify --useSSL -X -p 1444 -D "uid=admin" -w welcome1
dn: ds-recurring-task-id=NightlyBackup2,cn=Recurring Tasks,cn=Tasks
changetype: add
objectClass: top
objectClass: ds-task
objectClass: ds-recurring-task
objectClass: ds-task-backup
description: Nightly backup at 2 AM
ds-backup-location: bak
ds-recurring-task-id: NightlyBackup2
ds-recurring-task-schedule: 00 02 * * *
ds-task-class-name: org.opends.server.tasks.BackupTask
ds-task-id: NightlyBackup2
ds-task-state: RECURRING
