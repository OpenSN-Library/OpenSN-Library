#!/bin/bash

sudo docker run --rm -it \
    --privileged=true \
    --network host \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e DOCKER_HOST=unix:///var/run/docker.sock \
    -e MODE=master \
    -e INTERFACE=ens160 \
    satellite_emulator/node-daemon