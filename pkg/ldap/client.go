/*
	Copyright 2021 ForgeRock AS.
*/
// Package ldap provides ldap client access to our DS deployment. Used to manage tasks, set passwords, etc.
package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

// DSConnection parameters for managing the DS ldap service
type DSConnection struct {
	URL      string
	DN       string
	Password string
	ldap     *ldap.Conn
}

type LdapObject struct {
	DN              string
	CreateTimeStamp string
	ObjectClass     []string
	// For simplicity we have a map of single valued string attributes.
	// For our needs we don't need multi valued or non string attributes.
	StringAttrs map[string]string
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

// GetEntry gets an user ldap entry by its UID. The search starts under ou=identities
func (ds *DSConnection) getEntryByUID(uid string) (*LdapObject, error) {

	req := ldap.NewSearchRequest("ou=identities",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uid="+uid+")",
		[]string{"dn", "cn", "uid", "mail", "displayName", "givenName", "sn", "description", "createTimestamp", "objectClass"}, // A list attributes to retrieve
		nil)

	res, err := ds.ldap.Search(req)
	if err != nil {
		return nil, err
	}

	if len(res.Entries) != 1 {
		return nil, fmt.Errorf("User not found or more than one entry matched")
	}

	ra := res.Entries[0]
	sa := make(map[string]string)

	for _, a := range ra.Attributes {
		sa[a.Name] = a.Values[0]
	}

	return &LdapObject{
		DN:              ra.DN,
		CreateTimeStamp: ra.GetAttributeValue("createTimeStamp"),
		ObjectClass:     ra.GetAttributeValues("objectClass"),
		StringAttrs:     sa,
	}, nil
}

// BindPassword tries to bind as the DN with the password. This is used to test the password to see if we need to change it.
// Return nil if the password is OK, err otherwise
func (ds *DSConnection) BindPassword(DN, password string) error {
	//ds.Log.V(2).Info("ldap client - BIND", "DN", DN)
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
	//ds.Log.V(2).Info("ldap client - update password", "DN", DN)
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	return err
}

// Create a sample user. Used for testing, but could be used in the future for creating admin service accounts.
func (ds *DSConnection) AddEntry(obj *LdapObject) error {
	req := ldap.NewAddRequest(obj.DN, nil)
	req.Attribute("objectClass", obj.ObjectClass)
	// add all the single valued string attributes
	for k, v := range obj.StringAttrs {
		req.Attribute(k, []string{v})
	}

	err := ds.ldap.Add(req)
	return err
}

// Get an Entry by its DN. attrs is a list of attributes to return. objectClass and createTimeStamp are always returned.
func (ds *DSConnection) GetEntryByDN(DN string, attrs []string) (*LdapObject, error) {
	attrs = append(attrs, "createTimeStamp")
	attrs = append(attrs, "objectClass")

	req := ldap.NewSearchRequest(DN, ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=*)", attrs, nil)

	res, err := ds.ldap.Search(req)
	if err != nil {
		return nil, err
	}

	if len(res.Entries) != 1 {
		return nil, fmt.Errorf("Object with dn %s not found", DN)
	}
	ra := res.Entries[0]
	sa := make(map[string]string)

	for _, a := range ra.Attributes {
		//fmt.Printf("attrs=%s:%s\n", a.Name, a.Values[0])
		sa[a.Name] = a.Values[0]
	}

	return &LdapObject{
		DN:              DN,
		CreateTimeStamp: ra.GetAttributeValue("createTimeStamp"),
		ObjectClass:     ra.GetAttributeValues("objectClass"),
		StringAttrs:     sa,
	}, nil

}

// Delete an LDAP entry specified by dn
func (ds *DSConnection) DeleteEntry(dn string) error {
	dr := ldap.DelRequest{DN: dn}
	return ds.ldap.Del(&dr)
}

// calculate the DN of a purge recurring task
func purgeTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-purge,cn=Recurring Tasks,cn=Tasks"
}

// calculate the DN of a task
func backupTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-backup,cn=Recurring Tasks,cn=Tasks"
}

// Create a task in DS. Currently suports only purge and backup. If we need more tasks, consider refactoring this to be
// more generic
// This is currently not used - but may be in the future if we return to direct DS to S3 backups
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

// getMonitorData returns cn=monitor data. We use this for status updates.
// todo: What kinds of data do we want to monitor?
func (ds *DSConnection) getMonitorData() error {

	req := ldap.NewSearchRequest("cn=monitor",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 100, 0, false,
		"(objectclass=*)",
		[]string{},
		nil)

	res, err := ds.ldap.Search(req)

	fmt.Printf("%d monitoring entries found\n", len(res.Entries))

	// res.PrettyPrint(2)

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
