#!/bin/bash -er
docker rmi satellite-monitor
docker builder prune
bash build_dockerfile.sh