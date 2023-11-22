import etcd,redis
import time,json
from loguru import logger
from global_var import *

'''
type Instance struct {
    InstanceID        string
    Name              string
    Type              string
    PositionChangable bool
    ContainerID       string
    NodeID            uint32
    Namespace         string
    LinksID           []string
    Extra             map[string]string
}

Type = SATELLITE = "Satellite"

Extra Has
TLE_0 -> TLE_LINE0
TLE_1 -> TLE_LINE1
TLE_2 -> TLE_LINE2

'''


def ParseEtcdResult(info : dict[bytes,bytes]):
    add_key : list[bytes] = []
    del_key : list[bytes] = []
    for remote_key in info.keys():
        if remote_key not in MovingInstances.keys():
            add_key.append(remote_key)
    for local_key in MovingInstances.keys():
        if local_key not in info.keys():
            del_key.append(local_key)
    
    for key in del_key:
        del MovingInstances[key]
    for key in add_key:
        obj_map = json.loads(info[key])
        if obj_map["type"] == ["Satellite"]:
            inst = Instance(
                    key.decode(),
                    [
                        obj_map["extra"]["TLE_0"],
                        obj_map["extra"]["TLE_1"],
                        obj_map["extra"]["TLE_1"]
                    ]
                )
            MovingInstances[key]=inst
        

def watch_instance():
    etcd_client = etcd.client(host=ETCD_ADDR,port=ETCD_PORT)
    redis_client = redis.Redis(host=REDIS_ADDR,port=REDIS_PORT,password=REDIS_PASSWORD)
    while True:
        try:
            etcd_client.watch(NODE_INS_WATCH_KEY)
            infos = redis_client.get(NODE_INS_INFO_KEY)
            ParseEtcdResult(infos)
        except etcd.EtcdKeyNotFound as e:
            logger.warning("Node Instance Infomation Crashed or InstanceManager Not Init, Waiting...")
            time.sleep(1)
        except Exception as e:
            logger.error("Read Instance Modules Error %s"%str(e))
            time.sleep(10)