#!/usr/bin/env python3

# Script to generate sample LDIF for ds idrepo user store.
# generates two ldif files: identities.ldif and relationships.ldif


import uuid
import datetime
import sys
import random

numusers = 1000000

realm = "alpha"

person_suffix = "ou=user,o=" + realm + ",o=root,ou=identities"
meta_suffix = "ou=usermeta,o=" + realm + ",o=root,ou=identities"
relationship_suffix = "ou=relationships,dc=openidm,dc=forgerock,dc=io"


def telnum():
	val = random.randrange(9999999999, 9999999999999)

	return val


# objectClass: fr-ext-attrs


def person(seq, puid, muid, ruid):
	first = "f" + str(seq)
	last = "l" + str(seq)
	cn = first + " " + last
	uid = "user." + str(seq)
	tel = telnum()

	print(f"dn: fr-idm-uuid={puid},{person_suffix}")
	print('''objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
objectClass: iplanet-am-user-service
objectClass: devicePrintProfilesContainer
objectClass: deviceProfilesContainer
objectClass: kbaInfoContainer
objectClass: fr-idm-managed-user-explicit
objectClass: forgerock-am-dashboard-service
objectClass: inetuser
objectClass: iplanet-am-auth-configuration-service
objectClass: iplanet-am-managed-person
objectClass: iPlanetPreferences
objectClass: oathDeviceProfilesContainer
objectClass: pushDeviceProfilesContainer
objectClass: sunAMAuthAccountLockout
objectClass: sunFMSAML2NameIdentifier
objectClass: webauthnDeviceProfilesContainer
objectClass: fr-idm-hybrid-obj
inetUserStatus: active
userPassword: T35tr0ck123*''')
	print(f"mail: user.{seq}@example.com")
	print(f"fr-idm-uuid: {puid}")
	print(f"cn: {cn}")
	print(f"givenName: {first}")
	print(f"sn: {last}")
	print(f"telephoneNumber: {tel}")
	print(f"uid: {uid}")

	print('fr-idm-managed-user-meta: {"firstResourceCollection":"managed/%s_user", \
"firstResourceId":"%s", "firstPropertyName":"_meta", \
"secondResourceCollection":"managed/%s_usermeta","secondResourceId":"%s", \
"secondPropertyName":null,"properties":null,"_rev":"0000000000000000","_id":"%s"}uid=%s,%s'  % (realm, puid, realm, muid, ruid, muid, meta_suffix) )
	print()


def usermeta(muid):
	cdate = datetime.datetime.now().strftime("%Y-%m-%dT%X.%fZ") # 2021-09-01T17:18:33.639825Z
												
	print(f"dn: uid={muid},{meta_suffix}")
	print('''objectClass: top
objectClass: uidObject
objectClass: fr-idm-generic-obj''')
	print('fr-idm-json: {"createDate":"%s","lastChanged":{"date":"%s"},"loginCount":0}' % (cdate, cdate))
	print(f"uid: {muid}")
	print()


def relationship(ruid, puid, muid, ldif):
	ldif.writelines(f"dn: uid={ruid},{relationship_suffix}\n")
	ldif.writelines('''objectClass: top
objectClass: uidObject
objectClass: fr-idm-relationship\n''')
	ldif.writelines(f"uid: {ruid}\n")
	ldif.writelines('fr-idm-relationship-json: {"firstResourceCollection":"managed/%s_user","firstResourceId":"%s", \
"firstPropertyName":"_meta","secondResourceCollection":"managed/%s_usermeta", \
"secondResourceId":"%s","secondPropertyName":null,"properties":null}\n' % (realm, puid, realm, muid))
	ldif.writelines("\n")

def branches():
	print('''dn: ou=identities
objectClass: top
objectClass: organizationalUnit

dn: o=root,ou=identities
objectClass: top
objectClass: organization
o: root

dn: o=%s,o=root,ou=identities
objectClass: top
objectClass: organization
o: alpha

dn: ou=user,o=%s,o=root,ou=identities
objectClass: top
objectClass: organizationalUnit
ou: user

dn: ou=usermeta,o=%s,o=root,ou=identities
objectClass: top
objectClass: organizationalUnit
ou: usermeta''' % (realm, realm, realm) )
	print()


def main():
	sys.stdout = open('identities.ldif', 'w')
	branches()

	with open('relationships.ldif', 'w') as ldif:
		for i in range(numusers):
			puid = uuid.uuid4()
			muid = uuid.uuid4()
			ruid = uuid.uuid4()

			person(i, puid, muid, ruid)
			usermeta(muid)

			relationship(ruid, puid, muid, ldif)
	ldif.close()
	sys.stdout.close()

if __name__ == "__main__":
    main()
