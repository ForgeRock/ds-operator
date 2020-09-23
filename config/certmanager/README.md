# DS Operator - prototype / work in progress

Experimental!!

## Goals

* Gain experience with a complex operator, to further refine requirements.
* Provide simpler and more robust deployment for directory services.
* Automate tasks that would need to be manually performed (enable backup, restore from a backup).
* Deploy mulitple DS instances without requiring multiple kustomize bases and overlays
* Enable optional features
  * L4 external load balancer for LDAP, HTTP
  * Run across different namespaces

## Design considerations

* Sensible defaults. For example, can we default the docker image?
* Provide "flavors" - CTS vs IdRepo ?
* Secrets Integration. Should the operator require secret-agent?
  * Suggest no - use secretRefs to refer to secrets. Make secret agent optional.
  * Have operator generate secrets itself if a secretRef is not provided?

Initial implementation:

* Mirror the current implementation. Deploy statefulset

Simplest possible deployment:

* metadata.name - name of DS statefulset
* spec.replicas - number of ds replicas
* image: cts or ds-idrepo (eventually these can be flavors)
* Storage size
* CPU / Memory requests

Example

```
# A small deployment using sane defaults where possible
apiVersion: directory.forgerock.com/v1alpha1
kind: DirectoryService
metadata:
  name: cts
spec:
  replicas: 3
  image: gcr.io/forgeops-public/ds-cts:latest
  diskSize: 100G
  resources:
    requests:
        memory: 512Mi
        cpu: 250m
    limits:
        memory: 1024Mi
```

Optional features

* Storage class configuration
* anti-affinity
* L4 load balancing


Questions:

* How do we integrate backup/restore? Some operators use a separate operator CRD for backups.  Seems a bit klunky.
* How do you get backup status. Is that part of the info returned by `describe` on the DS object?