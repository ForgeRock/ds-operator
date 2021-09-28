/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"
	"strconv"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

// PasswordCheckTimeSeconds is the number of seconds between password checks.
// This is to avoid overloading the directory with frequent password checks/changes
const PasswordCheckTimeSeconds = 300

// annotation we add to the CR to record when we last checked the passwords to see if they are correct in the diretory
const LastPasswordCheckAnnotation = "directory.forgerock.io/last-password-check"

/// Manages ldap service account passwords in Spec.Passwords
func (r *DirectoryServiceReconciler) updatePasswords(ctx context.Context, ds *directoryv1alpha1.DirectoryService, ldap *ldap.DSConnection) error {
	log := k8slog.FromContext(ctx)

	now := time.Now().Unix()
	annotations := ds.ObjectMeta.GetAnnotations()
	lastTime := annotations[LastPasswordCheckAnnotation]

	if lastTime != "" {
		lastTimeInt, err := strconv.ParseInt(lastTime, 10, 64)
		if err != nil {
			log.Error(err, "Failed to parse last password check time")
			// set to 0 so we force a pw change, and this will reset the annotation
			lastTimeInt = 0
		}
		elapsed := now - lastTimeInt

		//fmt.Printf("DEBUG - elapsed time since password check %d\n", elapsed)

		// if it is too soon, skip the check so we are nice to the directory
		if elapsed < PasswordCheckTimeSeconds {
			return nil
		}
	}

	// for each password...
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
		// First check to see if we even need to change the password. It may be fine in which case we can go to the next one.
		if ldap.BindPassword(key, pw) == nil {
			log.V(5).Info("Current password for DN does not need to be changed", "dn", key)
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
	annotations[LastPasswordCheckAnnotation] = strconv.FormatInt(now, 10)
	ds.ObjectMeta.SetAnnotations(annotations)

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
