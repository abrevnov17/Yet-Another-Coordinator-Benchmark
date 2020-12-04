#!/bin/sh

# create namespace
kubectl apply -f yac_namespace.json

# establish namespace context
kubectl config set-context --current --namespace=yac

# create deployment
kubectl apply -f coordinator_deployment.yaml

# create service
kubectl apply -f coordinator_service.yaml