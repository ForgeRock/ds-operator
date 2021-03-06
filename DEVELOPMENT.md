# Development notes - Directory Services Operator

## Installing the Operator from source


The operator is deployed from source with the following command:

```bash
make manifests
kustomize build  config/default | kubectl apply -f -
```

## Secrets

Required secrets are not generated yet. You must use secret agent
as documented in the [README](README.md)

You must provide at a minimum the following secret

* ds - keystore with the master keystore and pin.

## Development Workflow

Note:

While the directory needs to run in Kubernetes, it is much easier to develop the operator running locally (outside of the cluster):

```bash
# See below for dev mode explanation
export DEV_MODE=true
make install
make run
# In another window...
kubectl apply -f hack/ds.yaml
kubectl scale directoryservice/ds --replicas=2
kubectl delete -f hack/ds.yaml
```

When testing out of cluster, the controller on your desktop needs to open an ldap connection to the directory that is
running in Kubernetes.
The DEV_MODE variable  (see above) sets the operator connection to localhost:1636, instead of the Kubernetes
pod hostname (default.ds-idrepo-0.ds-idrepo.cluster.local, for example) In dev mode, port forward to the ds container:

```bash
kubectl port-forward ds-idrepo-0 1636
```

This allows the operator running on your desktop to communicate with the directory server. This is needed
for any LDAP functionality.


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
* experimental ds proxy

Updating the spec.image will update the statefulset and perform a rolling update. For example:

```bash
kubectl patch directoryservice/ds --type='json' \
   -p='[{"op": "replace", "path": "/spec/image", "value":"gcr.io/forgeops-public/ds-idrepo:2020.10.28-AlSugoDiNoci"}]'
```


## Optional Features (Future Roadmap)

* Snapshots using k8s snapshots when supported on all clouds (Kubernetes 1.18)
* Patching strategy on DS image updates
* SSL / client authentication
* STS updates? See https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/#updating-statefulsets
In 1.17, some sts settings can be updated: image, Resource req/limit, labels and annotations. Might be useful to adjust JVM memory and allow restart
* Alerts?
 * Disk full alerts?
* Tuning. How do we tune backend params, or is that a function of the docker image?
* cli tool that can run commands in ds. For example, running dsconfig / dsbackup commands.  cli can grab the admin creds to make this simpler.

## DS JIRAs to track

* https://bugster.forgerock.org/jira/browse/OPENDJ-7502
* https://bugster.forgerock.org/jira/browse/OPENDJ-7501
* https://bugster.forgerock.org/jira/browse/OPENDJ-7352
* https://bugster.forgerock.org/jira/browse/CLOUD-2666
* https://bugster.forgerock.org/jira/browse/OPENDJ-5308 -  Task backend via REST

## Implementation Notes

(Scratch Notes to implementers...)

Spec update: https://kubernetes.slack.com/archives/CAR30FCJZ/p1602800878040500?thread_ts=1602647971.012900&cid=CAR30FCJZ
"One safe pattern is to mutate the spec, then update (i.e. commit) the spec, then mutate the status, then commit the status."

cn=monitor - status we might want to use for the operator status:

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

## Release Process

* Uses goreleaser and cloudbuild

To run a test build,  issue `/gcbrun` in the PR comment
To cut a release, create a release in GitHub. This creates a tag, and starts the cloudbuild / goreleaser process.
