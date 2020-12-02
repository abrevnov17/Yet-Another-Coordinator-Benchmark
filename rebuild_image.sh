#!/bin/sh

# Rebuilds docker image
docker build -t abrevnov/yac:latest .
docker push abrevnov/yac:latest