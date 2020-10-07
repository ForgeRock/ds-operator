package controllers

import (
	"context"
	"fmt"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

/// Manages ldap service account passwords

func (r *DirectoryServiceReconciler) updatePasswords(ctx context.Context, ds *directoryv1alpha1.DirectoryService, svc *v1.Service) (ctrl.Result, error) {
	log := r.Log

	if ds.Status.ServiceAccountPasswordsUpdatedAt.Seconds != 0 {
		// TODO: Do we want to trigger a password update if a secret changes?
		log.Info("Service account passwords already updated, nothing to do")
		return ctrl.Result{}, nil
	}
	log.Info("Updating service account passwords")

	//  ds.warren.svc.cluster.local

	// TODO: is there a more reliable way of getting the service hostname?
	// url := fmt.Sprintf("ldap://%s.%s.svc.cluster.local:1389", svc.Name, svc.Namespace)
	// For local testing we need to run kube port-forward and localhost...
	url := fmt.Sprintf("ldap://localhost:1389")

	// lookup the admin password. Do we want to cache this?
	var adminSecret v1.Secret
	account := ds.Spec.AccountSecrets["uid=admin"]
	name := types.NamespacedName{Namespace: ds.Namespace, Name: account.SecretName}

	if err := r.Get(ctx, name, &adminSecret); err != nil {
		log.Error(err, "Can't find secret for the admin password", "secret", name)
		return ctrl.Result{}, nil
	}

	password := adminSecret.Data[account.Key]

	ldap := ldap.DSConnection{DN: "uid=admin", URL: url, Password: string(password[:])}

	if err := ldap.Connect(); err != nil {
		log.Info("Can't connect to ldap server, will try again later", "url", url, "err", err)
		return ctrl.Result{RequeueAfter: time.Second * 30}, err
	}
	defer ldap.Close()

	for key, account := range ds.Spec.AccountSecrets {
		// we skip any internal accounts as they are managed on boot by the ds image
		if key == "uid=admin" || key == "uid=monitor" {
			continue
		}
		log.Info("Updating account password", "dn", key, "secretName", account.SecretName)
		// get the secret
		pw, err := r.lookupSecret(ctx, ds.Namespace, ds.SecretNameForDN(key), key)
		if err != nil {
			log.Error(err, "Can't find secret containing the password", "account", key, "secretName", account.SecretName)
			return ctrl.Result{RequeueAfter: time.Second * 30}, err
		}
		if err := ldap.UpdatePassword(key, pw); err != nil {
			log.Error(err, "Failed to update the password", "dn", key)
			return ctrl.Result{RequeueAfter: time.Second * 30}, err
		}
		log.Info("Updated password", "dn", key)
	}
	// if we get to here we have updated the passwords OK
	ds.Status.ServiceAccountPasswordsUpdatedAt = metav1.Timestamp{Seconds: time.Now().Unix()}

	// todo: Update the status so we dont do this everytime
	return ctrl.Result{}, nil
}

// Lookup a secret value
func (r *DirectoryServiceReconciler) lookupSecret(ctx context.Context, namespace, name, key string) (string, error) {
	n := types.NamespacedName{Namespace: namespace, Name: name}
	var secret v1.Secret
	if err := r.Get(ctx, n, &secret); err != nil {
		return "", err
	}
	password := secret.Data[key]
	return string(password[:]), nil
}
