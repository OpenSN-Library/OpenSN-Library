import os
from instance import Instance

NODE_INDEX = int(os.getenv("NODE_INDEX"))
ETCD_ADDR = os.getenv("ETCD_ADDR")
ETCD_PORT = int(os.getenv("ETCD_PORT"))
REDIS_ADDR = os.getenv("REDIS_ADDR")
REDIS_PORT = int(os.getenv("REDIS_PORT"))
REDIS_PASSWORD = int(os.getenv("REDIS_PASSWORD"))

MovingInstances:dict[bytes,dict[bytes,Instance]] = {}
# etcd keys

NODE_INS_INFO_KEY = "node_%d_instances"%NODE_INDEX # dir
NODE_NS_LIST_KEY = "/node_%d/ns_list"%NODE_INDEX
NODE_NS_POS_KEY_PREFIX = "/node_%d/"%NODE_INDEX