#!/bin/bash

if  [ ! "$1" ] ;then
    gap=5
else
    gap=${1}
fi


echo "Start Container Start Left Monitor,Check Gap ${1} second"

while :
do
    left=`sudo docker ps -a | grep Created | wc | awk '{print $1}'`
    echo  "time:"`date`";left:${left}"

    if [${left}=="0"]; then
        break
    fi

    sleep(${gap})
    
done