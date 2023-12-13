#!/bin/bash -er
sudo docker rmi satellite-node
# docker builder prune
bash build_dockerfile.sh