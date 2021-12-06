#!/usr/bin/env bash
# Creates a minikube cluster for testing
# Note the csi hostpath driver is very unreliable. You often need to delete/recreate the cluster.

minikube delete
minikube start --memory 7gb --cpus 3
minikube addons enable  csi-hostpath-driver
minikube addons enable volumesnapshots

# Create the volume snapshot class used in the samples.
kubectl apply -f - <<EOF
# Apply this on Minikube to create the VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshotClass
metadata:
  name: ds-snapshot-class
driver: hostpath.csi.k8s.io
deletionPolicy: Delete
EOF

kubectl create -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
provisioner: hostpath.csi.k8s.io
reclaimPolicy: Delete
volumeBindingMode: Immediate
EOF


kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml

# kubectl apply -f hack/secrets.yaml


make install

