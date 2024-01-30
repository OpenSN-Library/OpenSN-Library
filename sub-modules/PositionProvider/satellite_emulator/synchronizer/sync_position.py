from etcd3 import Etcd3Client
from satellite_emulator.const.etcd_key import POSITION_LIST_KEY
from satellite_emulator.model.position import Position,position_from_json,position_to_json

def put_position(etcd_client:Etcd3Client,instance_id:str, position:Position):
    position_key = "%s/%s"%(POSITION_LIST_KEY,instance_id)
    etcd_client.put(position_key,position_to_json(position))

def get_position(etcd_client:Etcd3Client,instance_id) -> Position:
    position_key = "%s/%s"%(POSITION_LIST_KEY,instance_id)
    val,meta = etcd_client.get(position_key)
    return position_from_json(val)

def get_position_map(etcd_client:Etcd3Client) -> dict[str,Position]:
    ret: dict[int,Position] = {}
    resps = etcd_client.get_prefix(POSITION_LIST_KEY)
    for val,meta in resps:
        ret[meta.key.decode().split('/')[-1]] = position_from_json(val)
    return ret