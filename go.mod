module github.com/ForgeRock/ds-operator

go 1.16

require (
	github.com/go-ldap/ldap/v3 v3.4.1
	github.com/go-logr/logr v1.2.2
	github.com/kubernetes-csi/external-snapshotter/client/v4 v4.2.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.23.3
	sigs.k8s.io/controller-runtime v0.11.0
)
