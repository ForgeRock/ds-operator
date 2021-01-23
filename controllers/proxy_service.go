/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *DirectoryServiceReconciler) reconcileProxyService(ctx context.Context, ds *directoryv1alpha1.DirectoryService) (v1.Service, error) {
	// create or update the service
	var svc v1.Service
	proxyName := ds.Name + "-proxy"

	if !ds.Spec.Proxy.Enabled {
		err := r.Client.Get(ctx, types.NamespacedName{Name: proxyName, Namespace: ds.Namespace}, &svc)
		if k8sErrors.IsNotFound(err) {
			return v1.Service{}, nil
		}
		// if the object exists, check if we own it
		if err == nil {
			owner := metav1.GetControllerOf(&svc)
			// if we own the object, delete it
			if owner.APIVersion == directoryv1alpha1.GroupVersion.String() && owner.UID == ds.GetUID() {
				return v1.Service{}, r.Client.Delete(ctx, &svc, client.PropagationPolicy("Background"))
			}
			return v1.Service{}, nil
		}
		return v1.Service{}, err
	}

	svc.Name = proxyName
	svc.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
		r.Log.V(8).Info("CreateorUpdate proxy service", "svc", svc)

		var err error
		// does the service not exist yet?
		if svc.CreationTimestamp.IsZero() {
			err = createProxyService(ds, &svc)
			r.Log.V(8).Info("Setting ownerref for proxy service", "svc", svc.Name)
			_ = controllerutil.SetControllerReference(ds, &svc, r.Scheme)
		} else {
			// If the service exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			//err = updateService(&ds, &sts)
			r.Log.V(8).Info("TODO: Handle update of ds proxy service")
		}

		r.Log.V(8).Info("proxy svc after update/create", "svc", svc)
		return err
	})
	return svc, err

}

// Create the service for ds
func createProxyService(ds *directoryv1alpha1.DirectoryService, svc *v1.Service) error {
	proxyName := ds.Name + "-proxy"
	labels := createLabels(proxyName, map[string]string{
		"app.kubernetes.io/component": "ds-proxy",
	})
	svcTemplate := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: make(map[string]string),
			Name:        proxyName,
			Namespace:   ds.Namespace,
		},
		Spec: v1.ServiceSpec{
			// ClusterIP: "None", // headless service
			Selector: map[string]string{
				"app.kubernetes.io/name":     labels["app.kubernetes.io/name"],
				"app.kubernetes.io/instance": labels["app.kubernetes.io/instance"],
			},
			Ports: []v1.ServicePort{
				{
					Name: "tcp-admin",
					Port: 4444,
				},
				{
					Name: "tcp-ldap",
					Port: 1389,
				},
				{
					Name: "tcp-ldaps",
					Port: 1636,
				},
				{
					Name: "https",
					Port: 8443,
				},
				{
					Name: "http",
					Port: 8080,
				},
			},
		},
	}

	svcTemplate.DeepCopyInto(svc)
	return nil // todo: can this ever fail?
}
