module github.com/ForgeRock/ds-operator

go 1.15

require (
	github.com/go-ldap/ldap/v3 v3.2.4
	github.com/go-logr/logr v0.3.0
	github.com/kubernetes-csi/external-snapshotter/client/v3 v3.0.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.2
)
