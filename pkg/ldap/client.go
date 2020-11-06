// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
package ldap

import (
	"fmt"
	"time"

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

// Connect to LDAP server via admin credentials
func (ds *DSConnection) Connect() error {
	l, err := ldap.DialURL(ds.URL)

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

// UpdatePassword changes the password for the user identified by the DN. This is done as an administrative password change
// The old password is not required.
func (ds *DSConnection) UpdatePassword(DN, newPassword string) error {
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	return err
}

// GetBackupTaskStatus queries for the completed backup tasks for the given id
func (ds *DSConnection) GetBackupTaskStatus(id string) ([]dir.DirectoryBackupStatus, error) {

	// TODO: Search needs to order by recent date. We need server side sort controls for this
	// https://github.com/go-ldap/ldap/issues/290

	// current time minus 2 days
	t := time2DirectoryTimeString(time.Now().AddDate(0, 0, -2))
	query := fmt.Sprintf("(&(ds-recurring-task-id=%s-backup)(ds-task-scheduled-start-time>=%s))", id, t)

	// return 24 results. Too many results will clutter the status update
	req := ldap.NewSearchRequest("cn=Scheduled Tasks,cn=tasks",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 24, 0, false, query,
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

// DeleteBackupTask deletes a scheduled backup and purge tasks in DS.
// deletes the backup and the purge tasks in DS
// TODO: Check for Not found error code- which we can ignore and not consider an error
func (ds *DSConnection) DeleteBackupTask(id string) error {
	req := ldap.NewDelRequest(purgeTaskDN(id), []ldap.Control{})
	err1 := ds.ldap.Del(req)

	req2 := ldap.NewDelRequest(backupTaskDN(id), []ldap.Control{})
	err2 := ds.ldap.Del(req2)

	if err1 != nil || err2 != nil {
		return fmt.Errorf("%v %v", err1, err2)
	}
	return nil
}

func purgeTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-purge,cn=Recurring Tasks,cn=Tasks"
}

func backupTaskDN(id string) string {
	return "ds-recurring-task-id=" + id + "-backup,cn=Recurring Tasks,cn=Tasks"
}

// ScheduleBackup - create or update a backup and purge tasks
func (ds *DSConnection) ScheduleBackup(id string, d *dir.DirectoryBackup) error {
	// delete the existing task id.. Ignore any failed deletes..
	_ = ds.DeleteBackupTask(id)
	// schedule the backup task
	if err := ds.createTask(id, backupTaskDN(id), d.Cron, d.Path, "ds-task-backup", 0); err != nil {
		return err
	}
	// schedule the purge task
	return ds.createTask(id, purgeTaskDN(id), d.PurgeCron, d.Path, "ds-task-purge", d.PurgeHours)
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

// GetBackupTask reads the directory and gets the current state of the backup task.
func (ds *DSConnection) GetBackupTask(id string) (*dir.DirectoryBackup, error) {

	// filter for OR of either task DN
	filter := fmt.Sprintf("(|(ds-recurring-task-id=%s-backup)(ds-recurring-task-id=%s-purge))", id, id)

	req := ldap.NewSearchRequest("cn=Recurring Tasks,cn=Tasks",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{}, // return the default set of entries
		nil)
	res, err := ds.ldap.Search(req)
	//res.PrettyPrint(2)
	if err != nil {
		return nil, err
	}

	if len(res.Entries) <= 0 {
		return nil, fmt.Errorf("No Backup task found")
	}

	var d dir.DirectoryBackup

	d.Enabled = true // if there are backup tasks, it must be enabled

	for _, entry := range res.Entries {
		if entry.DN == purgeTaskDN(id) {
			d.PurgeCron = entry.GetAttributeValue("ds-recurring-task-schedule")
			hours := entry.GetAttributeValue("ds-task-purge-older-than")
			fmt.Sscanf(hours, "%d", &d.PurgeHours)
		} else if entry.DN == backupTaskDN(id) {
			d.Cron = entry.GetAttributeValue("ds-recurring-task-schedule")
			d.Path = entry.GetAttributeValue("ds-backup-location")
		} else {
			return &d, fmt.Errorf("Unexpected DN found %s", entry.DN)
		}

	}
	return &d, nil

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
