#!/bin/bash

container_ids=`sudo docker ps -a --format "table {{.Names}}" | grep -v "NAMES"`

for container_id in ${container_ids}
do 
    sudo docker exec ${container_id} ip route > ${container_id}.route
    echo "${container_id}:"
    cat ${container_id}.route
done