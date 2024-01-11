#!/bin/bash
links=`ip l | grep ns1 | awk '{print $2}' | awk '{split($1,arr,"@"); print arr[1]}'`
for link in ${links[@]} 
do
    sudo ip link del ${link}
done