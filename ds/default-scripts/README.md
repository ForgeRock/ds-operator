# Default Scripts

These life cycle scripts are defaults to be used if the user does not provide
their own script.  You can override these scripts by providing your own
scripts in a configmap. See the `spec.scriptConfigMapName` field in the DirectoryService CR.

The life cycle scripts are:

* `setup`: Called when the PVC data volume is empty. This should set the directory server up,
 create all backends, indexes and acis. The default script creates a complete "idrepo" and cts configuration suitable for running the ForgeOps CDK.
 * `backup`: Called by the DirectoryBackup Job. The Job will have a `/backup` pvc mounted to hold the backup files. The `/opt/opendj/data` will contain a clone (via snapshot) of the DS data. The backup script should perform any action needed to backup the DS data. The sample provided exports to LDIF format.
 * `restore`: Called by the DirectoryRestore Job. The Job will have a `/backup` pvc mounted that holds the data to be restored (ldif, dsbackup, etc.). The `/opt/opendj/data` directory will be mounted ready to be restored. The restore script should perform any action needed to restore the DS data. (ldif import, dsrestore). The  provided default sample imports from LDIF format. When the data restore is complete, the a volume snapshot is created of the data directory. This snapshot can be used to restore a cluster based on the snapshot.
 * `post-init`: If the user supplies a post-init script it will be called by the init container after the linking and index rebuilds are
 performed. Use this to add any new indexes before the server starts, or to issue an other dsconfig commands. The directory will be offline
 when this script is run.




