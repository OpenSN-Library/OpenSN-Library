#!/bin/bash
export NODE_ID=0 \
    ETCD_ADDR=10.134.148.56 \
    ETCD_PORT=4001 \
    REDIS_ADDR=10.134.148.56 \
    REDIS_PORT=6379 \
    REDIS_PASSWORD=123456 \
    && python3 main.py