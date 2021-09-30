//go:build integration
// +build integration

/*
	Copyright 2021 ForgeRock AS.
*/

// Package ldap provides ldap client access to our DS deployment. Used to manage users, tasks, etc.
// IMPORTANT NOTE:  This is in an *integration* test that requires a running ldap server. This test will not run standalone!

package ldap

import (
	"fmt"
	"testing"
)

// Set the Directory Admin password
const (
	PASSWORD = "R7yjdKNARusxFEv54I31i1NYQ9xoZvIt"
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
			entry, err := ds.getEntryByUID("am-identity-bind-account")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
			fmt.Printf("entry %v", entry)
			// When testing against DJ make sure to use a strong password that passes the policy (>8, special chars, upper/lower)
			err = ds.UpdatePassword("uid=am-identity-bind-account,ou=admins,ou=identities", "Password123!")
			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}

			// test user create / delete
			user := LdapObject{
				ObjectClass: []string{"inetOrgPerson", "organizationalPerson", "person", "top"},
				DN:          "uid=testuser,ou=people,ou=identities",
				StringAttrs: map[string]string{
					"uid":          "testuser",
					"cn":           "Test User",
					"sn":           "User",
					"mail":         "test@test.com",
					"userPassword": "Passw0rd!123",
					"givenName":    "Test",
				},
			}

			if err := ds.AddEntry(&user); err != nil {
				t.Errorf("Add Entry failed %v", err)
			}

			// read the user entry back
			u, err := ds.GetEntryByDN(user.DN, []string{"dn", "objectclass", "uid", "createTimeStamp"})

			if err != nil {
				t.Errorf("Get Entry failed %v", err)
			}
			fmt.Printf("%+v\n", u)

			// Delete the user
			if err := ds.DeleteEntry(u.DN); err != nil {
				t.Errorf("Delete Entry failed %v", err)
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
			if err := ds.getMonitorData(); err != nil {
				t.Errorf("Can't read monitoring data")
			}

		})
	}
}
