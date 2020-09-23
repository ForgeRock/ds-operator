 # DS Operator - experimental


## Design notes

* Secrets should be by reference?   `spec.secretRef`  references externally generated secrets
* Optional: If no secretRefs are provided, operator will generate random secrets and fill in the secretRefs


## What works now

* basic statefulset is created and comes up OK
* deletinng the CR properly cleans up the statefulset. PVC is left behind - which is a good thing

## What doesn't work

* No backup/restore yet
* No service definition or load balancer
* No status updates on the CR object
* No modification when the CR changes. Need to understand what changes in the CR can be reflected in the deployments
  * spec.Replicas should increase the number of sts replicas
  * spec.Image should update the base image
  