package main

import (
	"fmt"

	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
)

func main() {
	l := ldap.DSConnection{
		URL:      "ldap://localhost:1389",
		Password: "Ju1mXhL5lIzrYntHyli6Qnb67P39t7fK",
		DN:       "uid=admin",
	}
	if err := l.Connect(); err != nil {
		panic(err)
	}
	defer l.Close()

	user := ldap.User{
		DN:          "uid=testuser,ou=people,ou=identities",
		UID:         "testuser",
		CN:          "Test User",
		SN:          "User",
		Mail:        "test@test.com",
		Password:    "Passw0rd!123",
		Description: "Test User",
		DisplayName: "Test User",
		GivenName:   "Test",
	}

	if err := l.AddUser(&user); err != nil {
		panic(err)
	}

	// read the user back
	u, err := l.GetUser(user.UID)

	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", u)

	// Delete the user
	if err := l.DeleteEntry(u.DN); err != nil {
		panic(err)
	}

}
