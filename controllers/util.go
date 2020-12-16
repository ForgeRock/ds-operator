/*
	Copyright 2020 ForgeRock AS.
*/
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

// Add all the standard labels to the labels map, and return the new map
// instanceName is the unique instance (ds-idrepo, ds-cts, etc.)
// If the labels map is nil or empty, just return the standard labels
func createLabels(instanceName string, labels map[string]string) map[string]string {
	l := map[string]string{
		"app.kubernetes.io/managed-by": "ds-operator",
		"app.kubernetes.io/name":       LabelApplicationName,
		"app.kubernetes.io/instance":   instanceName,
		"app.kubernetes.io/part-of":    "forgerock",
	}
	if labels != nil {
		for k, v := range labels {
			l[k] = v
		}
	}
	return l
}
