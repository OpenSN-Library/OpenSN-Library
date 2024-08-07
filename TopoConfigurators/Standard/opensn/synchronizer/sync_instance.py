from etcd3 import Etcd3Client
from opensn.const.etcd_key import NODE_INST_KEY_TEMPLATE,\
    NODE_INST_RUNTIME_KEY_TEMPLATE,\
    NODE_INSTANCE_CONFIG_KEY_TEMPLATE
from opensn.model.instance import Instance,\
    instance_from_json,\
    instance_to_json

def put_instance(etcd_client:Etcd3Client,instance: Instance):
    instance_key = "%s/%s"%(NODE_INST_KEY_TEMPLATE%instance.node_index,instance.instance_id)
    etcd_client.put(instance_key,instance_to_json(instance))

def remove_instance(etcd_client:Etcd3Client,node_index,instance_id: str):
    instance_key = "%s/%s"%(NODE_INST_KEY_TEMPLATE%node_index,instance_id)
    etcd_client.delete(instance_key)

def get_instance(etcd_client:Etcd3Client,node_index,instance_id:str) -> Instance:
    instance_key = "%s/%s"%(NODE_INST_KEY_TEMPLATE%node_index,instance_id)
    val,meta = etcd_client.get(instance_key)
    return instance_from_json(val)

def get_instance_map(etcd_client:Etcd3Client,node_index: int) -> dict[str,Instance]:
    ret: dict[str,Instance] = {}
    base_key = NODE_INST_KEY_TEMPLATE%node_index
    resps = etcd_client.get_prefix(base_key)
    for val,meta in resps:
        ret[meta.key.decode().split('/')[-1]] = instance_from_json(val)
    return ret

def put_instance_config(etcd_client:Etcd3Client,node_index,instance_id:str,config_seq:str):
    instance_config_key = "%s/%s"%(NODE_INSTANCE_CONFIG_KEY_TEMPLATE%node_index,instance_id)
    etcd_client.put(instance_config_key,config_seq)

def put_instance_config_if_not_exist(etcd_client:Etcd3Client,node_index,instance_id:str,config_seq:str):
    instance_config_key = "%s/%s"%(NODE_INSTANCE_CONFIG_KEY_TEMPLATE%node_index,instance_id)
    etcd_client.put_if_not_exists(instance_config_key,config_seq)