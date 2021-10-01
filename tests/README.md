# DS Operator E2E Tests

## Testing Snapshots with a load.

### Prepare NS

```
❯ ./tests/bin/setup-ns test-snaps-with-load
+ kubectl create ns test-snaps-with-load
namespace/test-snaps-with-load created
+ kns='kubectl -n test-snaps-with-load'
+ kubectl -n test-snaps-with-load apply -f /home/max/projects/ds-operator/tests/bin/../secret_agent.yaml
secretagentconfiguration.secret-agent.secrets.forgerock.io/forgerock-sac created
+ kubectl -n test-snaps-with-load create configmap ldap-client --from-file=ldap=/home/max/projects/ds-operator/tests/bin/ldap --from-file=bench=/home/max/projects/ds-operator/tests/bin/run-bench
configmap/ldap-client created
+ extcode=1
+ [[ extcode -eq 0 ]]
+ kubectl -n test-snaps-with-load get secret ds-passwords -o 'jsonpath={.data.dirmanager\.pw}'
+ base64 --decode
htff1tMEwPy2eQGk1QVv79udpVaV2Ac9+ extcode=0
+ sleep 1
+ [[ extcode -eq 0 ]]
+ kubectl -n test-snaps-with-load apply -f /home/max/projects/ds-operator/tests/bin/../ds-test-client.yaml
deployment.apps/debug-ldap created
+ sleep 5
+ kubectl -n test-snaps-with-load rollout status deploy debug-ldap
deployment "debug-ldap" successfully rolled out
```
### Running a test

You'll need at least three shells to complete this:

#### Shell 1:
(note below shows an error rate but that's because this instance was restored so bench users already exist).
```
❯ k exec -it deploy/debug-ldap -- /opt/scripts/bench ds-idrepo-0.ds-idrepo
+ host=ds-idrepo-0.ds-idrepo
+ /opt/scripts/ldap make-user-on-host
IllegalArgumentException: BindRequestGenerator host name cannot be null at
Reject.java:269 Reject.java:119 LdapClientProvider.java:1287
LdapClientProvider.java:1718 LdapClientProvider.java:1688 Utils.java:843
Utils.java:826 LdapModify.java:180 Utils.java:787 Utils.java:764
LdapModify.java:98
+ cat
+ /opt/opendj/bin/addrate --hostname ds-idrepo-0.ds-idrepo --port 1636 --useSsl --trustAll --bindDn uid=admin --bindPassword IgrRYlEVy65SaJZwyFcl9Y8JXG39ZwPS --noPurge --noRebind --numConcurrentRequests 10 addrate.template
--------------------------------------------------------------------------------------
|     Throughput    |                 Response Time                |    Additional   |
|    (ops/second)   |                (milliseconds)                |    Statistics   |
|   recent  average |   recent  average    99.9%   99.99%  99.999% |  err/sec   Add% |
--------------------------------------------------------------------------------------
|   2774.5   2774.5 |    3.511    3.511    12.78    16.25    16.52 |   2774.5 100.00 |
|   4181.5   3477.7 |    2.350    2.813    11.80    15.99    16.52 |   4181.5 100.00 |
....
....
```

Wait about 30 mins and make sure you've landed a recent volumesnapshot, you chould have at least three. Go to shell 3.

```
^C

Now delete the DS deployment
k delete -f tests/ds-snapshots.yaml
k delete pvc --all
```

Now run the same commands as you did in shell2, and shell3. Verify the output is roughly the same and that DS recovers from a snapshot.

#### Shell 2

```
☸ dsoptests (max) in ds-operator on  test-snapshot-during-bench [$!?] via 🐹 v1.16.8 took 10s
❯ k apply -f tests/ds-snapshots.yaml
directoryservice.directory.forgerock.io/ds-idrepo created

☸ dsoptests (max) in ds-operator on  test-snapshot-during-bench [$!?] via 🐹 v1.16.8
❯ kubectl scale directoryservice/ds-idrepo --replicas=3
directoryservice.directory.forgerock.io/ds-idrepo scaled

☸ dsoptests (max) in ds-operator on  test-snapshot-during-bench [$!?] via 🐹 v1.16.8
❯ kubectl rollout status statefulset ds-idrepo
Waiting for 3 pods to be ready...
Waiting for 2 pods to be ready...
Waiting for 1 pods to be ready...
partitioned roll out complete: 3 new pods have been updated...
```
#### Shell 3
```
❯ k exec -it deploy/debug-ldap -- /opt/scripts/ldap count-ppl-on-host ds-idrepo-0.ds-idrepo
dn: ou=people,ou=identities
numsubordinates: 264967
```
