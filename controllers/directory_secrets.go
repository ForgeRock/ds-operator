package controllers

import (
	"math/rand"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create secret templates for secrets we need to create
// This iterates through the list of secrets, seeing which ones we own
// and need to create vs. those that we assume are already present
func createSecretTemplates(ds *directoryv1alpha1.DirectoryService) []v1.Secret {
	var secrets []v1.Secret

	for dn, accountSecret := range ds.Spec.AccountSecrets {
		if accountSecret.Create {
			// we own creating the secret
			secretTemplate := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      make(map[string]string),
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
