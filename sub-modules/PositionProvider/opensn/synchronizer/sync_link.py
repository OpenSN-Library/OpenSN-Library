from opensn.model.link import LinkBase,link_from_json,link_to_json,link_parameter_from_json,link_parameter_to_json
from opensn.const.etcd_key import NODE_LINK_KEY_TEMPLATE,NODE_LINK_PARAMETER_TEMPLATE
from etcd3 import Etcd3Client

def put_link(etcd_client:Etcd3Client,link: LinkBase):
    link_key = "%s/%s"%(NODE_LINK_KEY_TEMPLATE%link.end_infos[0].end_node_index,link.link_id)
    link_seq = link_to_json(link)
    etcd_client.put(link_key,link_seq)
    if link.end_infos[0].end_node_index != link.end_infos[1].end_node_index:
        link_key = "%s/%s"%(NODE_LINK_KEY_TEMPLATE%link.end_infos[1].end_node_index,link.link_id)
        etcd_client.put(link_key,link_seq)

def remove_link(etcd_client:Etcd3Client,node_index:int, link_id: str):
    link_key = "%s/%s"%(NODE_LINK_KEY_TEMPLATE%node_index,link_id)
    etcd_client.delete(link_key)

def get_link(etcd_client:Etcd3Client,node_index,link_id:str) -> LinkBase:
    link_key = "%s/%s"%(NODE_LINK_KEY_TEMPLATE%node_index,link_id)
    val,meta = etcd_client.get(link_key)
    return link_from_json(val)

def get_link_map(etcd_client:Etcd3Client,node_index: int) -> dict[str,LinkBase]:
    ret: dict[str,LinkBase] = {}
    base_key = NODE_LINK_KEY_TEMPLATE%node_index
    resps = etcd_client.get_prefix(base_key)
    for val,meta in resps:
        ret[meta.key.decode().split('/')[-1]] = link_from_json(val)
    return ret

def get_link_parameter(etcd_client:Etcd3Client,node_index:int,link_id:str):
    link_parameter_key = "%s/%s"%(NODE_LINK_PARAMETER_TEMPLATE%node_index,link_id)
    val,meta = etcd_client.get(link_parameter_key)
    return link_parameter_from_json(val)

def put_link_parameter(etcd_client:Etcd3Client,node_index:int,link_id:str,parameter: dict[str,int]):
    link_parameter_key = "%s/%s"%(NODE_LINK_PARAMETER_TEMPLATE%node_index,link_id)
    link_seq = link_parameter_to_json(parameter)
    etcd_client.put(link_parameter_key,link_seq)

