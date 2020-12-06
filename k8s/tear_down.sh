#!/bin/sh

# establishing context
kubectl config set-context --current --namespace=yac

# deleting deployment
kubectl delete deployment coordinator

# deleting service
kubectl delete service coordinator

# deleting namespace
kubectl delete namespaces yac

# restore namespace
kubectl config set-context --current --namespace=default