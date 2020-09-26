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

	// Name of secret that contains the ds passwords. Defaults to $Name-passwords. Must have keys: dirmanage.pw, monitor.pw
	// TODO: This should be generated if the secret is not found.
	SecretReferencePasswords string `json:"secretReferencePasswords,omitempty"`
	SecretReference          string `json:"secretReference,omitempty"`
}

// DirectoryServiceStatus defines the observed state of DirectoryService
type DirectoryServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`
}

// +kubebuilder:object:root=true

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
