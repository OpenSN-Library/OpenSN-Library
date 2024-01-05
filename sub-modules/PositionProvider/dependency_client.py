import etcd3,redis
from const_var import ETCD_ADDR,ETCD_PORT,REDIS_ADDR,REDIS_PASSWORD,REDIS_PORT
etcd_client = etcd3.client(host=ETCD_ADDR, port=ETCD_PORT)
redis_client = redis.Redis(host=REDIS_ADDR,port=REDIS_PORT,password=REDIS_PASSWORD)