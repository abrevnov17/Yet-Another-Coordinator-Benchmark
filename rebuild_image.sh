#!/bin/sh

# Rebuilds docker image
docker build -t abrevnov/yac-baseline:latest .
docker push abrevnov/yac-baseline:latest