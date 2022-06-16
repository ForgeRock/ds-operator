/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"
	"math/rand"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

// loop over all the secrets that we own, and create or update
// Note there may be DirectoryPasswords that are referenced (bring your own secrets use case), but we don't own them
func (r *DirectoryServiceReconciler) reconcileSecrets(ctx context.Context, ds *directoryv1alpha1.DirectoryService) error {
	log := k8slog.FromContext(ctx)

	// Loop through the spec.passwords - creating secrets as required
	for dn, accountSecret := range ds.Spec.Passwords {
		if accountSecret.Create {
			secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: ds.SecretNameForDN(dn), Namespace: ds.Namespace}}

			_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
				if secret.CreationTimestamp.IsZero() {
					log.V(8).Info("Created Secret", "secret", secret.ObjectMeta.Name)
					secret.ObjectMeta.Labels = createLabels(ds.Name, nil)
				}
				if _, ok := secret.Data[accountSecret.Key]; ! ok {
					log.V(8).Info("Updating Secret", "secret", secret.ObjectMeta.Name)
					secret.Data[accountSecret.Key] = []byte(randPassword(24))
				}
				return nil
			})
			if err != nil {
				return errors.Wrap(err, "unable to CreateOrUpdate Secret")
			}
		}
	}
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!$^#()-+<>")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randPassword(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
