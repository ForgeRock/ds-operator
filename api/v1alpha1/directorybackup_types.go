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

/*
 This is modeled after the Velero API.

 kubebuilder create api --group directory --version v1alpha1 --kind DirectoryBackup


*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DirectoryBackupSpec defines the desired state of DirectoryBackup
type DirectoryBackupSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// DirectoryPVCClaim is the PVC that contains the directory data. Make an array???
	ClaimToBackup string `json:"claimToBackup"`

	// Snapshot class name to use for all snapshots.
	VolumeSnapshotClassName string `json:"volumeSnapshotClassName"`

	// +kubebuilder:validation:Required
	VolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"volumeClaimSpec,required"`

	// Kubernetes resources assigned to the pod
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Docker Image for the directory server.
	Image string `json:"image"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Never;IfNotPresent;Always
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Certificates - needed for reading/writing encrypted data
	Certificates DirectoryCertificates `json:"certificates,required"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DirectoryBackup is the Schema for the directorybackups API
type DirectoryBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DirectoryBackupSpec `json:"spec,omitempty"`
	Status BackupStatus        `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DirectoryBackupList contains a list of DirectoryBackup
type DirectoryBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DirectoryBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DirectoryBackup{}, &DirectoryBackupList{})
}

// BackupStatus captures the current status of a backup.
type BackupStatus struct {

	// StartTimestamp records the time a backup was started.
	// The server's time is used for StartTimestamps
	// +optional
	// +nullable
	StartTimestamp *metav1.Time `json:"startTimestamp,omitempty"`

	// CompletionTimestamp records the time a backup was completed.
	// Completion time is recorded even on failed backups.
	// Completion time is recorded before uploading the backup object.
	// The server's time is used for CompletionTimestamps
	// +optional
	// +nullable
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// Progress contains information about the backup's execution progress. Note
	// that this information is best-effort only -- if Velero fails to update it
	// during a backup for any reason, it may be inaccurate/stale.
	// +optional
	// +nullable
	//Progress *BackupProgress `json:"progress,omitempty"`
}
