#!/bin/bash

container_ids=`sudo docker ps -a --format "table {{.Names}}" | grep -v "NAMES"`

for container_id in ${container_ids}
do 
    pid=`sudo docker inspect ${container_id} -f '{{.State.Pid}}'`
    sudo nsenter --net=/proc/${pid}/ns/net ip route > ${container_id}.route
    echo "${container_id}:"
    cat ${container_id}.route
done