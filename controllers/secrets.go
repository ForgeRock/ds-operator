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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// loop over all the secrets that we own, and create or update
// Note there may be DirectoryPasswords that are referenced (bring your own secrets use case), but we don't own them
func (r *DirectoryServiceReconciler) reconcileSecrets(ctx context.Context, ds *directoryv1alpha1.DirectoryService) error {

	// Loop through the spec.passwords - creating secrets as required
	for _, secret := range createSecretTemplates(ds) {
		_, err := ctrl.CreateOrUpdate(ctx, r.Client, &secret, func() error {
			if secret.CreationTimestamp.IsZero() {
				r.Log.V(8).Info("Created Secret", "secret", secret)
				_ = controllerutil.SetControllerReference(ds, &secret, r.Scheme)
			} else {
				// The secret already exists... Do we want to update it?
				r.Log.V(8).Info("TODO: Update secret", "secret", secret)
			}
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "unable to CreateOrUpdate Secret")
		}
	}

	return nil
}

// Create secret templates for secrets we need to create
// This iterates through the list of secrets, seeing which ones we own
// and need to create vs. those that we assume are already present
func createSecretTemplates(ds *directoryv1alpha1.DirectoryService) []v1.Secret {
	var secrets []v1.Secret

	for dn, accountSecret := range ds.Spec.Passwords {
		if accountSecret.Create {
			// we own creating the secret
			secretTemplate := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      createLabels(ds.Name, nil),
					Annotations: make(map[string]string),
					Name:        ds.SecretNameForDN(dn),
					Namespace:   ds.Namespace,
				},
				Data: map[string][]byte{
					accountSecret.Key: []byte(randPassword(24)),
				},
			}
			secrets = append(secrets, secretTemplate)
		}
	}
	return secrets
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
