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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackupPVC struct {
	Name             string `json:"name"`
	Size             string `json:"size"`
	StorageClassName string `json:"storageClassName"`
}

// DirectoryBackupSpec defines the desired state of DirectoryBackup
type DirectoryBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	BackupPVC BackupPVC `json:"backupPVC,omitempty"`

	// DirectoryPVCClaim is the PVC that contains the directory data. Make an array???
	ClaimToBackup string `json:"claimToBackup"`

	// Snapshot class name to use for all snapshots.
	VolumeSnapshotClassName string `json:"volumeSnapshotClassName"`
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

// BackupPhase is a string representation of the lifecycle phase
// of a Velero backup.
// +kubebuilder:validation:Enum=New;FailedValidation;InProgress;Completed;PartiallyFailed;Failed;Deleting
type BackupPhase string

const (
	// BackupPhaseNew means the backup has been created but not
	// yet processed by the BackupController.
	BackupPhaseNew BackupPhase = "New"

	// BackupPhaseFailedValidation means the backup has failed
	// the controller's validations and therefore will not run.
	BackupPhaseFailedValidation BackupPhase = "FailedValidation"

	// BackupPhaseInProgress means the backup is currently executing.
	BackupPhaseInProgress BackupPhase = "InProgress"

	// BackupPhaseCompleted means the backup has run successfully without
	// errors.
	BackupPhaseCompleted BackupPhase = "Completed"

	// BackupPhasePartiallyFailed means the backup has run to completion
	// but encountered 1+ errors backing up individual items.
	BackupPhasePartiallyFailed BackupPhase = "PartiallyFailed"

	// BackupPhaseFailed means the backup ran but encountered an error that
	// prevented it from completing successfully.
	BackupPhaseFailed BackupPhase = "Failed"

	// BackupPhaseDeleting means the backup and all its associated data are being deleted.
	BackupPhaseDeleting BackupPhase = "Deleting"
)

// BackupStatus captures the current status of a backup.
type BackupStatus struct {

	// FormatVersion is the backup format version, including major, minor, and patch version.
	// +optional
	FormatVersion string `json:"formatVersion,omitempty"`

	// Expiration is when this Backup is eligible for garbage-collection.
	// +optional
	// +nullable
	Expiration *metav1.Time `json:"expiration,omitempty"`

	// Phase is the current state of the Backup.
	// +optional
	Phase BackupPhase `json:"phase,omitempty"`

	// ValidationErrors is a slice of all validation errors (if
	// applicable).
	// +optional
	// +nullable
	ValidationErrors []string `json:"validationErrors,omitempty"`

	// StartTimestamp records the time a backup was started.
	// Separate from CreationTimestamp, since that value changes
	// on restores.
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

	// Warnings is a count of all warning messages that were generated during
	// execution of the backup. The actual warnings are in the backup's log
	// file in object storage.
	// +optional
	Warnings int `json:"warnings,omitempty"`

	// Errors is a count of all error messages that were generated during
	// execution of the backup.  The actual errors are in the backup's log
	// file in object storage.
	// +optional
	Errors int `json:"errors,omitempty"`

	// Progress contains information about the backup's execution progress. Note
	// that this information is best-effort only -- if Velero fails to update it
	// during a backup for any reason, it may be inaccurate/stale.
	// +optional
	// +nullable
	//Progress *BackupProgress `json:"progress,omitempty"`
}
