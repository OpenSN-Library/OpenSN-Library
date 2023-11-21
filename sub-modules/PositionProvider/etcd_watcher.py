import etcd
import time
from loguru import logger
from global_var import *

def ParseEtcdResult(res: etcd.EtcdResult):
    

def watch_instance():
    client = etcd.Client(host=ETCD_ADDR,port=ETCD_PORT)
    node_info_key = "/Node_%d/Instances"
    node_watch_key = "/Node_%d/InstanceIDList"
    while True:

        try:
            client.watch(node_watch_key)
            infos = client.read(node_info_key,recursive=True)
            ParseEtcdResult(infos)
            break
        except etcd.EtcdKeyNotFound as e:
            logger.warning("Node Instance Infomation Crashed or InstanceManager Not Init, Waiting...")
            time.sleep(1)
        except Exception as e:
            logger.error("Read Instance Modules Error %s"%str(e))
            time.sleep(10)