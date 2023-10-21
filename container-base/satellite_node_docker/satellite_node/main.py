import logging
import time
import os
from broadcast_listener import satellite_simulator
from command_recevier import tcp_receiver
from multiprocessing import Process
from tools import get_ip_address, get_interface_address

logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s %(filename)s[line:%(lineno)d] %(levelname)s %(message)s',
                    datefmt='%Y-%m-%d %H:%M:%S',
                    filename='log.log',
                    filemode='w')

IP_TABLE_BASE = '/configuration/ip_tables/'

if __name__ == "__main__":
    broad_port = os.getenv("BROAD_PORT", None)
    if broad_port is None:
        logging.error('No broadcast port information')
        raise Exception('No broadcast port information')
    broad_port = int(broad_port)
    host_ip = os.getenv("HOST_IP", None)
    threshold_str = os.getenv("THRESHOLD", '0.95')
    threshold = float(threshold_str)
    if host_ip is None:
        raise Exception('No host ip information')
    node_id = os.getenv("NODE_ID", None)
    if node_id is None:
        logging.error('No node id information')
        raise Exception('No node id information')
    node_id_integer = int(node_id[5:])  # node_x
    procs = [
        Process(target=satellite_simulator, args=(broad_port, node_id, threshold,)),
        Process(target=tcp_receiver, args=(get_ip_address(),)),
        # Process(target=copy_the_configuration_file),
        # Process(target=call_netlink_exe, args=(node_id_integer,)),
    ]

        

    procs[0].start()  # we first stop the information broadcast
    procs[1].start()
    # procs[2].start()
    # procs[3].start()
    while True:
        for i in range(len(procs)):
            pass
        ip_tab = open(IP_TABLE_BASE+node_id+".conf",'w')
        interface_ip_list = get_interface_address()
        for item in interface_ip_list:
            ip_tab.write(item.split('/')[0] + "\n")
        time.sleep(60)
