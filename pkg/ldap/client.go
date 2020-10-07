// Package ldap provides ldap client access to our DS deployment. Used to manage users, etc.
package ldap

import (
	"fmt"

	ldap "github.com/go-ldap/ldap"
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
	//defer l.Close()

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
// This doesn't do much right now ... just searches for an entry
func (ds *DSConnection) getEntry(dn string) error {

	req := ldap.NewSearchRequest("ou=admins,ou=identities",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uid="+dn+")",
		[]string{"dn", "cn", "uid"}, // A list attributes to retrieve
		nil)

	res, err := ds.ldap.Search(req)

	for _, entry := range res.Entries {
		fmt.Printf("%s: %v cn=%s\n", entry.DN, entry.GetAttributeValue("uid"), entry.GetAttributeValue("cn"))
	}

	return err
}

// UpdatePassword changes the password for the user identified by the DN. This is done as an administrative password change
// The old password is not required.
func (ds *DSConnection) UpdatePassword(DN, newPassword string) error {
	req := ldap.NewPasswordModifyRequest(DN, "", newPassword)
	_, err := ds.ldap.PasswordModify(req)
	//fmt.Printf("res = %v gen pass=%v", res, res.GeneratedPassword)
	return err
}

// Close the ldap connection
func (ds *DSConnection) Close() {
	ds.ldap.Close()
}
