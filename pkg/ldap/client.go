// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
package ldap

import (
	"fmt"

	dir "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/go-ldap/ldap/v3"
)

// DSConnection parameters for managing the DS ldap service
type DSConnection struct {
	URL      string
	DN       string
	Password string
	ldap     *ldap.Conn
}

// BackupParams parameters for DS backups
type BackupParams struct {
	Cron string
	Path string
	ID   string
}

// Connect to LDAP server via admin credentials
func (ds *DSConnection) Connect() error {
	l, err := ldap.DialURL(ds.URL)

	if err != nil {
		return fmt.Errorf("Cant open ldap connection to %s using dn %s :  %s", ds.URL, ds.DN, err.Error())
	}

	err = l.Bind(ds.DN, ds.Password)

	fmt.Printf("Connection status = %v", err)

	if err != nil {
		defer l.Close()
		return fmt.Errorf("Cant bind ldap connection to %s wiht %s: %s ", ds.URL, ds.DN, err.Error())
	}
	ds.ldap = l
	return nil
}

// GetEntry get an ldap entry.
// This doesn't do much right now ... just searches for an entry. Just for testing
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

// UpdatePassword changes the password for the user identified by the DN. This is done as an administrative password change
// The old password is not required.
func (ds *DSConnection) UpdatePassword(DN, newPassword string) error {
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	return err
}

// GetBackupTask queries the backup task and returns the parameters
func (ds *DSConnection) GetBackupTask(id string) (*BackupParams, error) {

	req := ldap.NewSearchRequest("ds-recurring-task-id="+id+",cn=Recurring Tasks,cn=Tasks",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=ds-task-backup)",
		[]string{}, // return the default set of entries
		nil)

	res, err := ds.ldap.Search(req)
	if err != nil {
		// todo: Do we want to log this here?
		//fmt.Printf("ldap errror %v result is %v", err, res)
		return nil, err
	}

	var b BackupParams

	if len(res.Entries) == 1 {
		e := res.Entries[0]
		for _, attr := range e.Attributes {

			switch n := attr.Name; n {
			case "ds-recurring-task-schedule":
				b.Cron = attr.Values[0]
			case "ds-backup-location":
				b.Path = attr.Values[0]
			case "ds-recurring-task-id":
				b.ID = attr.Values[0]
			}
		}
	} else {
		return nil, fmt.Errorf("Expected exactly one entry got %d", len(res.Entries))
	}
	return &b, nil
}

// GetBackupTaskStatus queries for the completed backup tasks for the given id
func (ds *DSConnection) GetBackupTaskStatus(id string) ([]dir.DirectoryBackupStatus, error) {

	// Get the last 10 results
	// TODO: Search needs to order by recent date. We need server side sort controls for this
	// https://github.com/go-ldap/ldap/issues/290
	// In the interim we might have to put a > condition on the search filter.
	// Look for entries > time.now - 5 days
	req := ldap.NewSearchRequest("cn=Scheduled Tasks,cn=tasks",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 10, 0, false,
		"(ds-recurring-task-id="+id+")",
		//[]string{"ds-task-scheduled-start-time", "ds-task-completion-time", "ds-task-state"},
		[]string{},
		nil)

	var dstat []dir.DirectoryBackupStatus

	res, err := ds.ldap.Search(req)

	// An error 4 (sizes result exceeded) is expected - so ignore it
	if err != nil && !ldap.IsErrorAnyOf(err, ldap.LDAPResultSizeLimitExceeded) {
		return dstat, err
	}
	if len(res.Entries) == 0 {
		return dstat, nil
	}

	for _, e := range res.Entries {
		var item dir.DirectoryBackupStatus

		for _, attr := range e.Attributes {
			switch attr.Name {
			case "ds-task-scheduled-start-time":
				item.StartTime = attr.Values[0]
			case "ds-task-completion-time":
				item.EndTime = attr.Values[0]
			case "ds-task-state":
				item.Status = attr.Values[0]
			// todo: Capture status messages from ds
			// case "ds-task-log-messages":
			// 	item.Messages = append(item.Messages, attr.Values...)
			default:
				//fmt.Printf("att = %s", attr.Name)
			}
		}
		dstat = append(dstat, item)
	}
	return dstat, nil
}

// DeleteBackupSchedule deletes a scheduled backup. If the connection is OK, but the task does  not exist
// we still return ok.
func (ds *DSConnection) DeleteBackupSchedule(id string) error {
	task := "ds-recurring-task-id=" + id + ",cn=Recurring Tasks,cn=Tasks"
	req := ldap.NewDelRequest(task, []ldap.Control{})
	err := ds.ldap.Del(req)
	return err
}

// ScheduleBackup - create or update a backup task
// This can be done over 1389.
func (ds *DSConnection) ScheduleBackup(b *BackupParams) error {

	// See if the scheduled task already exists, and if it does, we dont attempt to reschedule
	oldparams, err := ds.GetBackupTask(b.ID)
	// If the search fails (err != nil) we still want to fall through and try to create the schedule
	// It might be the case that no schedule exists at all - which will have a fail with error 32
	if err == nil {
		if *oldparams == *b {
			return nil // params have not changed - nothing to do
		}
	}

	// delete the existing task id.. Ignore any failed deletes..
	err = ds.DeleteBackupSchedule(b.ID)

	// the dn needs to be unique for a recurring task
	req := ldap.NewAddRequest("ds-recurring-task-id="+b.ID+",cn=Recurring Tasks,cn=Tasks", []ldap.Control{})
	req.Attribute("objectclass", []string{"top", "ds-task", "ds-recurring-task", "ds-task-backup"})

	req.Attribute("description", []string{"backup auto scheduled by ds operator"})
	req.Attribute("ds-backup-location", []string{b.Path})
	req.Attribute("ds-recurring-task-id", []string{b.ID})
	req.Attribute("ds-task-id", []string{b.ID})
	req.Attribute("ds-task-state", []string{"RECURRING"})
	req.Attribute("ds-recurring-task-schedule", []string{b.Cron})
	req.Attribute("ds-task-class-name", []string{"org.opends.server.tasks.BackupTask"})
	// We set the storage props for all clouds - even if they are not used
	req.Attribute("ds-task-backup-storage-property", []string{
		"gs.credentials.path:/var/run/secrets/cloud-credentials-cache/gcp-credentials.json",
		"s3.keyId.env.var:AWS_ACCESS_KEY_ID", "s3.secret.env.var:AWS_SECRET_ACCESS_KEY",
		"az.accountName.env.var:AZURE_ACCOUNT_NAME", "az.accountKey.env.var:AZURE_ACCOUNT_KEY",
	})

	return ds.ldap.Add(req)
}

// Close the ldap connection
func (ds *DSConnection) Close() {
	ds.ldap.Close()
}
