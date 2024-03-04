#!/bin/bash
links=`ip link | grep "-" | grep qlen | awk '{print $2}' | awk '{split($1,arr,"@"); print arr[1]}'`
for link in ${links[@]} 
do
    sudo ip link del ${link}
done
links=`ip link | grep "state DOWN" | grep qlen | awk '{print $2}' | awk '{split($1,arr,"@"); print arr[1]}'`
for link in ${links[@]} 
do
    sudo ip link del ${link}
done
