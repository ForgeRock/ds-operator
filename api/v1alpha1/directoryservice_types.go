/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DirectoryServiceSpec defines the desired state of DirectoryService
type DirectoryServiceSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Docker Image for the directory server. Defaults if not provided?
	Image string `json:"image,omitempty"`
	// Replicas is the number of directory server instances to create
	Replicas *int32 `json:"replicas,required"`
	// Type of ds instance. Allowed - cts or idrepo? If allow setting the Image, we don't need a type?
	// DSType string `json:"dsType,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// The account secrets. The key is the DN of the secret (example, uid=admin)
	AccountSecrets map[string]DirectoryAccountSecrets `json:"accountSecrets"`
}

// DirectoryAccountSecrets is a reference to an account secret.
// The operator can set the passwords for accounts such as the uid=admin, uid=monitor and service accounts such as uid=idm-admin,ou=admins
type DirectoryAccountSecrets struct {
	// The name of a secret
	SecretName string `json:"secretName"`
	// The key within the secret
	Key string `json:"key"`
	// Create a random secret. Assumes no external secret manager is creating
	Create bool `json:"create,omitempty"`
}

// DirectoryServiceStatus defines the observed state of DirectoryService
type DirectoryServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Active                           []corev1.ObjectReference `json:"active,omitempty"`
	LastUpdate                       metav1.Timestamp         `json:"lastUpdateTime,omitempty"`
	CurrentReplicas                  *int32                   `json:"currentReplicas,omitempty"`
	ServiceAccountPasswordsUpdatedAt metav1.Timestamp         `json:"serviceAccountPasswordsUpdatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DirectoryService is the Schema for the directoryservices API
type DirectoryService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DirectoryServiceSpec   `json:"spec,omitempty"`
	Status DirectoryServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DirectoryServiceList contains a list of DirectoryService
type DirectoryServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DirectoryService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DirectoryService{}, &DirectoryServiceList{})
}

// SecretNameForDN looks up the secret name for the given dn (example, uid=admin)
// If the secret is one we generate, we prefix the name with metadata.name
func (ds *DirectoryService) SecretNameForDN(pathRef string) string {
	sec := ds.Spec.AccountSecrets[pathRef]
	if sec.Create {
		return ds.Name + "-" + sec.SecretName
	}
	return sec.SecretName
}
