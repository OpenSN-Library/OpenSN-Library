from instance import Instance
from global_var import MovingInstances,ETCD_ADDR,ETCD_PORT,NODE_NS_POS_KEY_PREFIX
from datetime import datetime
import time, json
import etcd

def calculate():
    client = etcd.client(host=ETCD_ADDR,port=ETCD_PORT)
    while True:
        now = datetime.now()
        for ns in MovingInstances.values():
            for ins in ns.values():
                ins.calculatePostion()
        for ns in MovingInstances.keys():
            ns_position_key = "%s%s/position"%(NODE_NS_POS_KEY_PREFIX,ns)
            client.write(ns_position_key,json.dumps(MovingInstances[ns]))
        now = datetime.now()
        time.sleep(now.microsecond/1000)

