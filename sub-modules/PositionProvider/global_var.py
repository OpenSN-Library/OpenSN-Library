import os
from instance import Instance

NODE_ID_STR = int(os.getenv("NODE_ID"))
ETCD_ADDR = os.getenv("ETCD_ADDR")
ETCD_PORT = int(os.getenv("ETCD_PORT"))

MovingInstances:dict[str,Instance] = {}
