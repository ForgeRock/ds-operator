 #!/usr/bin/env bash
# Sample of adding a new script

echo "DS Status"

status --hostname localhost --port 4444 \
    --bindDN uid=admin --bindPassword $(cat /var/run/secrets/admin/dirmanager.pw) \
    --trustAll

echo "Replication Status"

 dsrepl status --hostname localhost  --port 4444 \
    --bindDN uid=admin --bindPassword $(cat /var/run/secrets/admin/dirmanager.pw) \
    --trustAll \
    --showReplicas --showGroups --showChangeLogs

