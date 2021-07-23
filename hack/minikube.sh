#!/usr/bin/env bash
# Creates a minikube cluster for testing
# Note the csi hostpath driver is very unreliable. You often need to delete/recreate the cluster.

minikube delete
minikube start
minikube addons enable  csi-hostpath-driver
minikube addons enable volumesnapshots

kubectl apply -f hack/minikube-volume-snap-class.yaml

make install

