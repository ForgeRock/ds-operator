#!/usr/bin/env bash
# Creates a minikube cluster for testing
# Note the csi hostpath driver is very unreliable. You often need to delete/recreate the cluster.
minikube delete
minikube start  --memory=4096 --disk-size=40gb

minikube addons enable  csi-hostpath-driver
minikube addons enable volumesnapshots

kubectl apply -f secrets.yaml

cd ..
make install
