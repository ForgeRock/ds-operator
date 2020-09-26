 # DS Operator - experimental

## Hacking

Note: Secrets are not generated yet - so run another deployment to get secrets generated (e.g. ds-only, all, etc)

```bash
# We are not using webhooks just yet...
export ENABLE_WEBHOOKS="false"
make install
make run
kubectl apply -f hack/ds.yaml
kubectl delete -f hack/ds.yaml
```

## Design notes

* Secrets should be by reference:   `spec.secretRef`  references externally generated secrets
* Optional: If no secretRefs are provided, operator will generate random secrets and fill in the secretRefs


## What works now

* basic statefulset is created and comes up OK
* headless service created
* Deleting the CR properly cleans up the statefulset and service (owner refs are working OK). PVC is left behind - which is a good thing

## What doesn't work

* Pretty much everything ;-)
* No webhooks or validation
* No backup/restore yet
* No load balancer
* No status updates on the CR object
* No modification when the CR changes. Need to understand what changes in the CR can be reflected in the sts:
  * spec.Replicas should increase the number of sts replicas
  * spec.Image should update the base image and do a rolling update
  * Changes to spec.LoadBalancer?

## Optional Features

Are these valuable?

* The ability to specify a number of service accounts and assoicated secretRef. The idea is to have the operator
  set the password for the service account. The account should already exist as it is beyond the scope of the operator
  to manage service accounts, create ACIs, etc.

This would be helpful for [CLOUD-2178](https://bugster.forgerock.org/jira/browse/CLOUD-2178)

Example:

```yaml
serviceAccounts:
- name idm-admin
  dn:  uid=am-repo-user,in,ou=Admins,dc=am-config
  secretRef:  idrepo-password
```

In the case above, the operator can watch the secret, and rotate the password in DS if it changes
