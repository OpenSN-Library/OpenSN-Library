#!/bin/bash

container_ids=`sudo docker ps -a --format "table {{.Names}}" | grep -v "NAMES"`

for container_id in ${container_ids}
do 
    pid=`sudo docker inspect ${container_id} -f '{{.State.Pid}}'`
    sudo nsenter --net=/proc/${pid}/ns/net ip link > ${container_id}.link
    echo "${container_id}:"
    cat ${container_id}.link
done
