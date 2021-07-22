// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupPVC) DeepCopyInto(out *BackupPVC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupPVC.
func (in *BackupPVC) DeepCopy() *BackupPVC {
	if in == nil {
		return nil
	}
	out := new(BackupPVC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStatus) DeepCopyInto(out *BackupStatus) {
	*out = *in
	if in.StartTimestamp != nil {
		in, out := &in.StartTimestamp, &out.StartTimestamp
		*out = (*in).DeepCopy()
	}
	if in.CompletionTimestamp != nil {
		in, out := &in.CompletionTimestamp, &out.CompletionTimestamp
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStatus.
func (in *BackupStatus) DeepCopy() *BackupStatus {
	if in == nil {
		return nil
	}
	out := new(BackupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryBackup) DeepCopyInto(out *DirectoryBackup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryBackup.
func (in *DirectoryBackup) DeepCopy() *DirectoryBackup {
	if in == nil {
		return nil
	}
	out := new(DirectoryBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryBackup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryBackupList) DeepCopyInto(out *DirectoryBackupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DirectoryBackup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryBackupList.
func (in *DirectoryBackupList) DeepCopy() *DirectoryBackupList {
	if in == nil {
		return nil
	}
	out := new(DirectoryBackupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryBackupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryBackupSpec) DeepCopyInto(out *DirectoryBackupSpec) {
	*out = *in
	out.BackupPVC = in.BackupPVC
	out.Keystore = in.Keystore
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryBackupSpec.
func (in *DirectoryBackupSpec) DeepCopy() *DirectoryBackupSpec {
	if in == nil {
		return nil
	}
	out := new(DirectoryBackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryKeystores) DeepCopyInto(out *DirectoryKeystores) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryKeystores.
func (in *DirectoryKeystores) DeepCopy() *DirectoryKeystores {
	if in == nil {
		return nil
	}
	out := new(DirectoryKeystores)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryObject) DeepCopyInto(out *DirectoryObject) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryObject.
func (in *DirectoryObject) DeepCopy() *DirectoryObject {
	if in == nil {
		return nil
	}
	out := new(DirectoryObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryPasswords) DeepCopyInto(out *DirectoryPasswords) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryPasswords.
func (in *DirectoryPasswords) DeepCopy() *DirectoryPasswords {
	if in == nil {
		return nil
	}
	out := new(DirectoryPasswords)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryProxy) DeepCopyInto(out *DirectoryProxy) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryProxy.
func (in *DirectoryProxy) DeepCopy() *DirectoryProxy {
	if in == nil {
		return nil
	}
	out := new(DirectoryProxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryProxyStatus) DeepCopyInto(out *DirectoryProxyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryProxyStatus.
func (in *DirectoryProxyStatus) DeepCopy() *DirectoryProxyStatus {
	if in == nil {
		return nil
	}
	out := new(DirectoryProxyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryRestore) DeepCopyInto(out *DirectoryRestore) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryRestore.
func (in *DirectoryRestore) DeepCopy() *DirectoryRestore {
	if in == nil {
		return nil
	}
	out := new(DirectoryRestore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryRestore) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryRestoreList) DeepCopyInto(out *DirectoryRestoreList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DirectoryRestore, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryRestoreList.
func (in *DirectoryRestoreList) DeepCopy() *DirectoryRestoreList {
	if in == nil {
		return nil
	}
	out := new(DirectoryRestoreList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryRestoreList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryRestoreSpec) DeepCopyInto(out *DirectoryRestoreSpec) {
	*out = *in
	out.Keystore = in.Keystore
	out.RestorePVC = in.RestorePVC
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryRestoreSpec.
func (in *DirectoryRestoreSpec) DeepCopy() *DirectoryRestoreSpec {
	if in == nil {
		return nil
	}
	out := new(DirectoryRestoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryRestoreStatus) DeepCopyInto(out *DirectoryRestoreStatus) {
	*out = *in
	if in.StartTimestamp != nil {
		in, out := &in.StartTimestamp, &out.StartTimestamp
		*out = (*in).DeepCopy()
	}
	if in.CompletionTimestamp != nil {
		in, out := &in.CompletionTimestamp, &out.CompletionTimestamp
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryRestoreStatus.
func (in *DirectoryRestoreStatus) DeepCopy() *DirectoryRestoreStatus {
	if in == nil {
		return nil
	}
	out := new(DirectoryRestoreStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryService) DeepCopyInto(out *DirectoryService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryService.
func (in *DirectoryService) DeepCopy() *DirectoryService {
	if in == nil {
		return nil
	}
	out := new(DirectoryService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryServiceList) DeepCopyInto(out *DirectoryServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DirectoryService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryServiceList.
func (in *DirectoryServiceList) DeepCopy() *DirectoryServiceList {
	if in == nil {
		return nil
	}
	out := new(DirectoryServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DirectoryServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryServiceSpec) DeepCopyInto(out *DirectoryServiceSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Passwords != nil {
		in, out := &in.Passwords, &out.Passwords
		*out = make(map[string]DirectoryPasswords, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Keystore = in.Keystore
	out.TrustStore = in.TrustStore
	out.Snapshots = in.Snapshots
	in.Proxy.DeepCopyInto(&out.Proxy)
	out.MultiCluster = in.MultiCluster
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryServiceSpec.
func (in *DirectoryServiceSpec) DeepCopy() *DirectoryServiceSpec {
	if in == nil {
		return nil
	}
	out := new(DirectoryServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectoryServiceStatus) DeepCopyInto(out *DirectoryServiceStatus) {
	*out = *in
	if in.Active != nil {
		in, out := &in.Active, &out.Active
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.CurrentReplicas != nil {
		in, out := &in.CurrentReplicas, &out.CurrentReplicas
		*out = new(int32)
		**out = **in
	}
	out.ProxyStatus = in.ProxyStatus
	out.SnapshotStatus = in.SnapshotStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectoryServiceStatus.
func (in *DirectoryServiceStatus) DeepCopy() *DirectoryServiceStatus {
	if in == nil {
		return nil
	}
	out := new(DirectoryServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectorySnapshotSpec) DeepCopyInto(out *DirectorySnapshotSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectorySnapshotSpec.
func (in *DirectorySnapshotSpec) DeepCopy() *DirectorySnapshotSpec {
	if in == nil {
		return nil
	}
	out := new(DirectorySnapshotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MultiCluster) DeepCopyInto(out *MultiCluster) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MultiCluster.
func (in *MultiCluster) DeepCopy() *MultiCluster {
	if in == nil {
		return nil
	}
	out := new(MultiCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestorePVC) DeepCopyInto(out *RestorePVC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestorePVC.
func (in *RestorePVC) DeepCopy() *RestorePVC {
	if in == nil {
		return nil
	}
	out := new(RestorePVC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnapshotStatus) DeepCopyInto(out *SnapshotStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnapshotStatus.
func (in *SnapshotStatus) DeepCopy() *SnapshotStatus {
	if in == nil {
		return nil
	}
	out := new(SnapshotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustStore) DeepCopyInto(out *TrustStore) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustStore.
func (in *TrustStore) DeepCopy() *TrustStore {
	if in == nil {
		return nil
	}
	out := new(TrustStore)
	in.DeepCopyInto(out)
	return out
}
