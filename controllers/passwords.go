/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PasswordCheckTimeSeconds is the number of seconds between password checks. The first time through will trigger an immediate check.
// This is to avoid overloading the directory.
const PasswordCheckTimeSeconds = 300

/// Manages ldap service account passwords in Spec.Passwords
func (r *DirectoryServiceReconciler) updatePasswords(ctx context.Context, ds *directoryv1alpha1.DirectoryService, ldap *ldap.DSConnection) error {
	log := r.Log

	now := time.Now().Unix()
	elapsed := now - ds.Status.ServiceAccountPasswordsUpdatedTime

	// if it is to soon, skip the check so we are nice to the directory
	if elapsed < PasswordCheckTimeSeconds {
		return nil
	}
	// for each password
	for key, account := range ds.Spec.Passwords {
		// skip any internal accounts as they are managed on boot by the ds image
		if key == "uid=admin" || key == "uid=monitor" {
			continue
		}
		log.V(5).Info("Updating service account password", "dn", key, "secretName", account.SecretName, "key", account.Key)
		// get the secret containing the password
		pw, err := r.lookupSecret(ctx, ds.Namespace, ds.SecretNameForDN(key), account.Key)
		if err != nil {
			log.Error(err, "Can't find secret containing the password", "account", key, "secretName", account.SecretName)
			return err
		}
		// First check to see if we even need to change the password. It may be fine
		if ldap.BindPassword(key, pw) == nil {
			log.V(5).Info("Current password for DN is OK", "dn", key)
			continue
		}
		// Bind above failed, so lets try to change it:
		if err := ldap.UpdatePassword(key, pw); err != nil {
			log.Error(err, "Failed to update the password", "dn", key)
			return err
		}
		log.Info("Updated password", "dn", key)
	}
	// if we get to here we have updated all passwords OK
	ds.Status.ServiceAccountPasswordsUpdatedTime = now

	return nil
}

// Lookup a secret value.
func (r *DirectoryServiceReconciler) lookupSecret(ctx context.Context, namespace, name, key string) (string, error) {
	n := types.NamespacedName{Namespace: namespace, Name: name}
	var secret v1.Secret
	if err := r.Get(ctx, n, &secret); err != nil {
		return "", err
	}
	password := secret.Data[key]
	return string(password), nil
}
