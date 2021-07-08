/*
	Copyright 2020 ForgeRock AS.
*/
// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/go-logr/logr"
)

// DSConnection parameters for managing the DS ldap service
type DSConnection struct {
	URL      string
	DN       string
	Password string
	ldap     *ldap.Conn
	Log      logr.Logger
}

// Connect to LDAP server via admin credentials
func (ds *DSConnection) Connect() error {

	l, err := ldap.DialURL(ds.URL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))

	if err != nil {
		return fmt.Errorf("Cant open ldap connection to %s using dn %s :  %s", ds.URL, ds.DN, err.Error())
	}

	err = l.Bind(ds.DN, ds.Password)

	if err != nil {
		defer l.Close()
		return fmt.Errorf("Cant bind ldap connection to %s wiht %s: %s ", ds.URL, ds.DN, err.Error())
	}
	ds.ldap = l
	return nil
}

// GetEntry get an ldap entry.
// This doesn't do much right now ... just searches for an entry. Just for testing and to provide an example
func (ds *DSConnection) getEntry(dn string) (*ldap.Entry, error) {

	req := ldap.NewSearchRequest("ou=admins,ou=identities",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uid="+dn+")",
		[]string{"dn", "cn", "uid"}, // A list attributes to retrieve
		nil)

	res, err := ds.ldap.Search(req)
	if err != nil {
		return nil, err
	}

	// just for info...
	for _, entry := range res.Entries {
		fmt.Printf("%s: %v cn=%s\n", entry.DN, entry.GetAttributeValue("uid"), entry.GetAttributeValue("cn"))
	}

	return res.Entries[0], err
}

// BindPassword tries to bind as the DN with the password. This is used to test the password to see if we need to change it.
// Return nil if the password is OK, err otherwise
func (ds *DSConnection) BindPassword(DN, password string) error {
	ds.Log.V(2).Info("ldap client - BIND", "DN", DN)
	// get a new connection. We cant do this with th existing connection as it would unbind us from the admin account..
	tldap, err := ldap.DialURL(ds.URL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	defer tldap.Close()
	if err != nil {
		return err
	}
	return tldap.Bind(DN, password)
}

// UpdatePassword changes the password for the user identified by the DN. This is done as an administrative password change
// The old password is not required.
func (ds *DSConnection) UpdatePassword(DN, newPassword string) error {
	ds.Log.V(2).Info("ldap client - update password", "DN", DN)
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	return err
}

func purgeTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-purge,cn=Recurring Tasks,cn=Tasks"
}

func backupTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-backup,cn=Recurring Tasks,cn=Tasks"
}

// Create a task in DS. Currently suports only purge and backup. If we need more tasks, consider refactoring this to be
// more generic
func (ds *DSConnection) createTask(taskID string, taskDN string, cron string, backupPath string, taskObjClass string, purgeHours int32) error {
	req := ldap.NewAddRequest(taskDN, []ldap.Control{})
	req.Attribute("objectclass", []string{"top", "ds-task", "ds-recurring-task", taskObjClass})
	req.Attribute("description", []string{"task auto scheduled by ds-operator"})
	req.Attribute("ds-backup-location", []string{backupPath})
	//req.Attribute("ds-recurring-task-id", []string{taskID})
	req.Attribute("ds-task-id", []string{taskID})
	req.Attribute("ds-task-state", []string{"RECURRING"})
	req.Attribute("ds-recurring-task-schedule", []string{cron})

	if taskObjClass == "ds-task-purge" {
		req.Attribute("ds-task-class-name", []string{"org.opends.server.tasks.BackupPurgeTask"})
		h := fmt.Sprintf("%dh", purgeHours)
		fmt.Printf("hours=%s", h)
		req.Attribute("ds-task-purge-older-than", []string{h})
	} else {
		req.Attribute("ds-task-class-name", []string{"org.opends.server.tasks.BackupTask"})
	}

	// We set the storage props for all clouds - even if they are not used
	req.Attribute("ds-task-backup-storage-property", []string{
		"gs.credentials.path:/var/run/secrets/cloud-credentials-cache/gcp-credentials.json",
		"s3.keyId.env.var:AWS_ACCESS_KEY_ID", "s3.secret.env.var:AWS_SECRET_ACCESS_KEY",
		"az.accountName.env.var:AZURE_ACCOUNT_NAME", "az.accountKey.env.var:AZURE_ACCOUNT_KEY",
	})

	return ds.ldap.Add(req)
}

// GetMonitorData returns cn=monitor data. We use thi for status updates.
// todo: What kinds of data do we want?
func (ds *DSConnection) GetMonitorData() error {

	req := ldap.NewSearchRequest("cn=monitor",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 100, 0, false,
		"(objectclass=*)",
		[]string{},
		nil)

	res, err := ds.ldap.Search(req)

	res.PrettyPrint(2)

	return err
}

// Close the ldap connection
func (ds *DSConnection) Close() {
	ds.ldap.Close()
}

const timeFormatSpec = "20060102150500"

func directoryTime2Time(dsTime string) (time.Time, error) {
	return time.Parse(timeFormatSpec, dsTime)
}

func time2DirectoryTimeString(t time.Time) string {
	return t.Format(timeFormatSpec)
}
