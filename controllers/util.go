package controllers

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// todo: cant embeded this in other struct literals.
func makeMeta(name, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Name:        name,
		Namespace:   namespace,
	}
}
