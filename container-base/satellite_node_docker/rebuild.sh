#!/bin/bash -er
docker rmi satellite-node
docker builder prune
bash build_dockerfile.sh