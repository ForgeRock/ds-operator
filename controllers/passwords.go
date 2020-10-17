package controllers

import (
	"context"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

/// Manages ldap service account passwords

func (r *DirectoryServiceReconciler) updatePasswords(ctx context.Context, ds *directoryv1alpha1.DirectoryService, ldap *ldap.DSConnection) error {
	log := r.Log

	if ds.Status.ServiceAccountPasswordsUpdatedAt.Seconds != 0 {
		// TODO: Do we want to trigger a password update if a secret changes?
		log.Info("Service account passwords already updated, nothing to do")
		return nil
	}
	log.Info("Updating service account passwords")

	// for each password
	for key, account := range ds.Spec.Passwords {
		// skip any internal accounts as they are managed on boot by the ds image
		if key == "uid=admin" || key == "uid=monitor" {
			continue
		}
		log.Info("Updating service account password", "dn", key, "secretName", account.SecretName)
		// get the secret with the password
		pw, err := r.lookupSecret(ctx, ds.Namespace, ds.SecretNameForDN(key), key)
		if err != nil {
			log.Error(err, "Can't find secret containing the password", "account", key, "secretName", account.SecretName)
			return err
		}
		if err := ldap.UpdatePassword(key, pw); err != nil {
			log.Error(err, "Failed to update the password", "dn", key)
			return err
		}
		log.Info("Updated password", "dn", key)
	}
	// if we get to here we have updated the passwords OK
	ds.Status.ServiceAccountPasswordsUpdatedAt = metav1.Timestamp{Seconds: time.Now().Unix()}

	// todo: Update the status so we dont do this everytime
	return nil
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
