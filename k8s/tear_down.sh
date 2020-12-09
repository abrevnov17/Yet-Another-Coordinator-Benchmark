#!/bin/sh

# establishing context
kubectl config set-context --current --namespace=yac-baseline

# deleting deployment
kubectl delete deployment coordinator

# deleting service
kubectl delete service coordinator

# deleting namespace
kubectl delete namespaces yac-baseline

# restore namespace
kubectl config set-context --current --namespace=default