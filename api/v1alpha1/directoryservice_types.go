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

	// Docker Image for the directory server.
	Image string `json:"image,required"`
	// Replicas is the number of directory server instances to create
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,required"`
	// Type of ds instance. Allowed - cts or idrepo? If allow setting the Image, we don't need a type?
	// DSType string `json:"dsType,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// The account secrets. The key is the DN of the secret (example, uid=admin)
	Passwords map[string]DirectoryPasswords `json:"passwords"`
	// Keystore references
	Keystores DirectoryKeystores `json:"keystores,omitempty"`

	// Backup
	Backup DirectoryBackup `json:"backup,omitempty"`
	// Restore
	Restore DirectoryRestore `json:"restore,omitempty"`
}

// DirectoryPasswords is a reference to account secrets that contain passwords for the directory.
// The operator can set the passwords for accounts such as the uid=admin, uid=monitor and service accounts such as uid=idm-admin,ou=admins
type DirectoryPasswords struct {
	// The name of a secret
	SecretName string `json:"secretName"`
	// The key within the secret
	Key string `json:"key"`
	// Create a random secret if true. Otherwise assumes the secret already exists
	Create bool `json:"create,omitempty"`
}

// DirectoryKeystores provides a reference to the keystore secrets
type DirectoryKeystores struct {
	// The name of a secret containing the keystore
	// +kubebuilder:default:=ds
	KeyStoreSecretName   string `json:"keyStoreSecretName,required"`
	TrustStoreSecretName string `json:"trustStoreSecretName,omitempty"`
}

// DirectoryBackup defines how and where to backup DS to
type DirectoryBackup struct {
	Enabled    bool   `json:"enabled,required"`
	Path       string `json:"path,required"`
	Cron       string `json:"cron,required"`
	SecretName string `json:"secretName,omitempty"`
}

// DirectoryRestore defines how to restore a new directory from a backup
type DirectoryRestore struct {
	Enabled    bool   `json:"enabled,required"`
	Path       string `json:"path,required"`
	SecretName string `json:"secretName,omitempty"`
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
	BackupStatus                     []DirectoryBackupStatus  `json:"backupStatus,omitempty"`
}

// DirectoryBackupStatus provides the status of the backup
// NOTE: This does not work, because we are mixing schema (backup info and replica info above)
// TODO: Can we model this a separate resource, so kubectl describe does the right thing?
// +kubebuilder:printcolumn:name="Start Time",type="string",JSONPath=".status.backupStatus.startTime",description="Backup Start time"
// +kubebuilder:printcolumn:name="End Time",type="string",JSONPath=".status.backupStatus.startTime",description="Backup End time"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.backupStatus.endTime",description="Backup Start time"
type DirectoryBackupStatus struct {
	// note DS returns these as string values. For status is ok
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status    string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentReplicas

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
	sec := ds.Spec.Passwords[pathRef]
	if sec.Create {
		return ds.Name + "-" + sec.SecretName
	}
	return sec.SecretName
}
