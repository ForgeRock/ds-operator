#!/usr/bin/env bash

cd /home/forgerock

(cd /ldif && /home/forgerock/generate_ldif.py )

echo "Generated ldif"
ls -lh /ldif

PW=/var/run/secrets/ds/dirmanager.pw

ldapmodify --bindDn uid=admin --bindPasswordFile $PW \
    --hostname ds-idrepo-0.ds-idrepo --port 1389 \
    --continueOnError \
    --numConnections 10 \
    -f /ldif/identities.ldif


ldapmodify --bindDn uid=admin --bindPasswordFile $PW \
    --hostname ds-idrepo-0.ds-idrepo --port 1389 \
    --continueOnError \
    --numConnections 10 \
    -f /ldif/relationships.ldif

while true; do

sleep 1000
   
done
