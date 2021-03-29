/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"testing"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	apps "k8s.io/api/apps/v1"
)

func Test_createDSStatefulSet(t *testing.T) {
	type args struct {
		ds  *directoryv1alpha1.DirectoryService
		sts *apps.StatefulSet
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "STS volume claim DataSource gets populated with the snapshot if provided",
			args: args{ds: &directoryv1alpha1.DirectoryService{
				Spec: directoryv1alpha1.DirectoryServiceSpec{Storage: "5Gi", InitializeFromSnapshotName: "snap-ds-1"},
			}, sts: &apps.StatefulSet{}},
			wantErr: false},
		{name: "STS volume claim DataSource is NOT populated when the snapshot path is not provided",
			args: args{ds: &directoryv1alpha1.DirectoryService{
				Spec: directoryv1alpha1.DirectoryServiceSpec{Storage: "5Gi"},
			}, sts: &apps.StatefulSet{}},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createDSStatefulSet(tt.args.ds, tt.args.sts); (err != nil) != tt.wantErr {
				t.Errorf("createDSStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.ds.Spec.InitializeFromSnapshotName != "" {
				if tt.args.sts.Spec.VolumeClaimTemplates[0].Spec.DataSource.Name != tt.args.ds.Spec.InitializeFromSnapshotName {
					t.Errorf("The volume claim did not get populated as expected")
				}
			} else {
				if tt.args.sts.Spec.VolumeClaimTemplates[0].Spec.DataSource != nil {
					t.Errorf("The volume claim data source should NOT be populated %+v", tt.args.sts.Spec.VolumeClaimTemplates)
				}
			}

		})
	}
}
