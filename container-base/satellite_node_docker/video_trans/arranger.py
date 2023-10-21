
import os
from loguru import logger

IP_TABLE_BASE = '/configuration/ip_tables'
# IP_TABLE_BASE = '/home/satellite-2/Workspace/distributed_simulation/satellite-source-routing/configuration/ip_tables'

ip_table = {}
ip_list = []
list_arrange_index = 0
def refresh_ip_list():
    global ip_table,ip_list
    ip_list = []
    files = os.listdir(IP_TABLE_BASE)
    for f in files:
        ip_file = open(os.path.join(IP_TABLE_BASE,f),'r')
        node_id = f.split('.')[0]
        ips = ip_file.readlines()
        logger.info("Read" + str(ips))
        node_ip_list = []
        for item in ips:
            node_ip_list.append(item.split('\n')[0])
        ip_table[node_id] = node_ip_list
        ip_list.append(node_ip_list)
        ip_file.close()
    return ip_table

def get_compress_ip(const_compress_ip: str):
    # global list_arrange_index, ip_list
    # ans = ip_list[list_arrange_index % len(ip_list)]
    # list_arrange_index += 1
    # return ans
    return [const_compress_ip]

if __name__ == "__main__":
    dc = refresh_ip_list()
    print(dc)