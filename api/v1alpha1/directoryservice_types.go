/*

Copyright 2021 ForgeRock AS.


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

// DirectoryServiceSpec defines the desired state of DirectoryService
type DirectoryServiceSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	PodTemplate DirectoryPodTemplate `json:"podTemplate"`

	// Replicas is the number of directory server instances to create
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,required"`

	// The account secrets. The key is the DN of the secret (example, uid=admin)
	Passwords map[string]DirectoryPasswords `json:"passwords"`

	// Snapshots
	Snapshots DirectorySnapshotSpec `json:"snapshots,omitempty"`
	// Proxy configurations
	Proxy DirectoryProxy `json:"proxy,omitempty"`

	// +kubebuilder:validation:Optional
	// The name of a configmap to mount on /opt/opendj/scripts
	// Optional - if not provided no mount will be performed
	ScriptConfigMapName string `json:"scriptConfigMapName,omitempty"`
}

// DirecotoryPodTemplate provides the common configuration for all three CRDs
type DirectoryPodTemplate struct {
	// Docker Image for the directory server.
	Image string `json:"image,required"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Never;IfNotPresent;Always
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Kubernetes resources assigned to the pod
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Certificates needed for direcotory operation.
	Certificates DirectoryCertificates `json:"certificates"`
	// +kubebuilder:validation:Required
	VolumeClaimSpec corev1.PersistentVolumeClaimSpec `json:"volumeClaimSpec,required"`

	// +kubebuilder:validation:Optional
	// The name of a configmap to mount on /opt/opendj/scripts
	// Optional - if not provided no mount will be performed
	ScriptConfigMapName string `json:"scriptConfigMapName,omitempty"`

	Env []corev1.EnvVar `json:"env,omitempty"`

	// Name of the volumesnapshot class used in any snapshot operation
	// +kubebuilder:validation:Required
	VolumeSnapshotClassName string `json:"volumeSnapshotClassName,ommitempty"`
}

// DirectorySnapshotSpec defines how to take Volume Snapshots

type DirectorySnapshotSpec struct {
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,required"`
	// +kubebuilder:default:=30
	PeriodMinutes int32 `json:"periodMinutes,required"`
	// +kubebuilder:default:=10
	SnapshotsRetained int32 `json:"snapshotsRetained,required"`
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

// DirectoryCertificates required for operation of the directory server
type DirectoryCertificates struct {
	// +kubebuilder:default:=ds-master-keypair
	MasterSecretName string `json:"masterSecretName"`
	// +kubebuilder:default:=ds-ssl-keypair
	SSLSecretName string `json:"sslSecretName"`
	// +kubebuilder:default:="ds-ssl-keypair"
	TruststoreSecretName string `json:"truststoreSecretName"`
}

// DirectoryKeystores provides a reference to the keystore secrets
type DirectoryKeystores struct {
	// The name of a secret containing the keystore
	// +kubebuilder:default:=ds
	SecretName string `json:"secretName,required"`
}

// TrustStore defines a CA key pair
type TrustStore struct {
	// The name of a secret
	SecretName string `json:"secretName,required"`
	KeyName    string `json:"keyName,required"`
	// Create a random secret if true. Otherwise assumes the secret already exists
	// Not currently supported
	Create bool `json:"create,omitempty"`
}

// DirectoryServiceStatus defines the observed state of DirectoryService
type DirectoryServiceStatus struct {
	// +optional
	Active                             []corev1.ObjectReference `json:"active,omitempty"`
	CurrentReplicas                    *int32                   `json:"currentReplicas,omitempty"`
	ServiceAccountPasswordsUpdatedTime int64                    `json:"serviceAccountPasswordsUpdatedTime,omitempty"`
	ServerMessage                      string                   `json:"serverMessage,omitempty"`
	ProxyStatus                        DirectoryProxyStatus     `json:"proxyStatus,omitempty"`
	SnapshotStatus                     SnapshotStatus           `json:"snapshotStatus,omitempty"`
}

type SnapshotStatus struct {
	LastSnapshotTimeStamp int64 `json:"lastSnapshotTimeStamp"`
}

// DirectoryProxyStatus defines the observed state of DirectoryService Proxy
type DirectoryProxyStatus struct {
	Replicas      int32  `json:"replicas,omitempty"`
	ReadyReplicas int32  `json:"readyReplicas,omitempty"`
	ServerMessage string `json:"serverMessage,omitempty"`
}

// DirectoryProxy defines the settings of the directory proxy
type DirectoryProxy struct {
	Enabled bool `json:"enabled,required"`
	// Docker Image for the directory server.
	Image string `json:"image,required"`
	// Replicas is the number of directory server proxy instances to create
	// +kubebuilder:validation:Maximum:=8
	// +kubebuilder:validation:Minimum:=0
	Replicas int32 `json:"replicas,required"`
	// PrimaryGroupID specifies the group of servers the ds proxy should recognize as primary
	// If no value is provided, all available directory servers will be considered to be primary
	PrimaryGroupID string                      `json:"primaryGroupId,omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentReplicas
// +kubebuilder:resource:shortName=ds

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

// MultiCluster enables MCS and configures identifiers for multiple multi-cluster solutions
type MultiCluster struct {
	// +kubebuilder:default:=false
	McsEnabled bool `json:"mcsEnabled,omitempty"`
	// ClusterTopology is a comma separate string of identifiers for each cluster e.g. "europe,us"
	// +kubebuilder:validation:required
	ClusterTopology string `json:"clusterTopology"`
	// +kubebuilder:validation:required
	ClusterIdentifier string `json:"clusterIdentifier"`
}

func init() {
	SchemeBuilder.Register(&DirectoryService{}, &DirectoryServiceList{})
}

// SecretNameForDN looks up the secret name for the given dn (example, uid=admin)
func (ds *DirectoryService) SecretNameForDN(pathRef string) string {
	sec := ds.Spec.Passwords[pathRef]
	return sec.SecretName
}
