import logging
import os
from socket import *
from subprocess import Popen, PIPE


def get_ip_address():
    p = Popen("ifconfig eth0 | grep 'inet' | awk '{print $2}'", shell=True, stdout=PIPE)
    data = p.stdout.read()
    ip = str(data[:-1], encoding='UTF-8')
    if ":" in ip:
        ip = ip.split(":")[-1]
    return ip

def get_interface_address():
    ip_list_raw = Popen("ip a | grep -E '([0-9]+\.){3}[0-9]+/[0-9]+'  | awk '{print $2}'", shell=True, stdout=PIPE)
    ip_list = str(ip_list_raw.stdout.read()[:-1], encoding='UTF-8').split('\n')
    if len(ip_list) > 2 :
        return ip_list[2:]
    return []

def get_node_id_from_env():
    """
    get the node id from the environment variable
    :return:
    """
    node_id = os.getenv("NODE_ID", None)
    if node_id is None:
        logging.error('No node id information')
        raise Exception('No node id information')
    node_id_integer = int(node_id[5:])  # node_x
    return node_id_integer


if __name__ == '__main__':
    print(get_ip_address())
