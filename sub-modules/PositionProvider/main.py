from global_var import *
import ephem
import threading
from datetime import datetime
from instance import Instance
from etcd_watcher import watch_instance
import etcd


if __name__ == "__main__":
    # watch_thread = threading.Thread(target=watch_instance)
    # watch_thread.run()
    client = etcd.Client(host=ETCD_ADDR,port=ETCD_PORT)
    client.write("/aaa/ccc",3,ttl=3)
    client.write("/aaa/bbb/qqq",4,ttl=3)
    client.write("/aaa/bbb/ppp",5,ttl=3)
    res = client.read("/aaa",recursive=True)
    print(res)