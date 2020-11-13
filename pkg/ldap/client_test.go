// +build integration

/*
	Copyright 2020 ForgeRock AS.
*/

// Package ldap provides ldap client access to our DS deployment. Used to manage users, tasks, etc.
// IMPORTANT NOTE:  This is in an integration test that requires a running ldap server. This test will not run standalone

package ldap

import (
	"fmt"
	"testing"

	dir "github.com/ForgeRock/ds-operator/api/v1alpha1"
)

// Set the Directory Admin password
const (
	PASSWORD = "Ij3uu7QBZn1u5Nt8"
	LDAP_URL = "ldaps://localhost:1636"
)

func TestDSConnection_Connect_test(t *testing.T) {
	type fields struct {
		url      string
		dn       string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Pick the password up from a file
		{"localhost test", fields{LDAP_URL, "uid=admin", PASSWORD}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DSConnection{
				DN:       tt.fields.dn,
				Password: tt.fields.password,
				URL:      tt.fields.url,
			}
			if err := ds.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DSConnection.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer ds.Close()
			entry, err := ds.getEntry("am-identity-bind-account")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
			fmt.Printf("entry %v", entry)
			// When testing against DJ make sure to use a strong password that passes the policy (>8, special chars, upper/lower)
			err = ds.UpdatePassword("uid=am-identity-bind-account,ou=admins,ou=identities", "Password123!")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
		})
	}
}

func TestDSAdmin(t *testing.T) {
	type fields struct {
		url      string
		dn       string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Pick the password up from a file
		{"localhost test", fields{LDAP_URL, "uid=admin", PASSWORD}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DSConnection{
				DN:       tt.fields.dn,
				Password: tt.fields.password,
				URL:      tt.fields.url,
			}

			if err := ds.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("DSConnection.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer ds.Close()
			// if err := ds.GetMonitorData(); err != nil {
			// 	t.Errorf("Can't read monitoring data")
			// }

			_ = ds.DeleteBackupTask("ds")

			dsb := dir.DirectoryBackup{
				Enabled:    true,
				Cron:       "55 * * * *",
				PurgeCron:  "*/5 * * * *",
				Path:       "/var/tmp/backup",
				PurgeHours: 12,
			}

			// try to schedule a backup
			if err := ds.ScheduleBackup("ds", &dsb); err != nil {
				t.Errorf("Backup schedule failed %v", err)
			}

			var dsb2 *dir.DirectoryBackup
			var e error

			if dsb2, e = ds.GetBackupTask("ds"); e != nil {
				t.Errorf("Could not get backup tasks %v", e)
			}

			fmt.Printf("\nGot %+v\n", dsb2)

		})
	}
}
