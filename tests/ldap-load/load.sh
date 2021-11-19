#!/usr/bin/env bash

cd /home/forgerock

(cd /ldif && /home/forgerock/generate_ldif.py )

date

echo "Generated ldif"
ls -lh /ldif

date

PW=/var/run/secrets/ds/dirmanager.pw

time ldapmodify --bindDn uid=admin --bindPasswordFile $PW \
    --hostname ds-idrepo-0.ds-idrepo --port 1389 \
    --continueOnError \
    --numConnections 10 \
    -f /ldif/identities.ldif


time ldapmodify --bindDn uid=admin --bindPasswordFile $PW \
    --hostname ds-idrepo-0.ds-idrepo --port 1389 \
    --continueOnError \
    --numConnections 10 \
    -f /ldif/relationships.ldif

echo done
while true; do

sleep 1000

date

done
