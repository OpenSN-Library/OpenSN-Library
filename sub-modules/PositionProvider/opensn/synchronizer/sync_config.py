from etcd3 import Etcd3Client
from opensn.model.emulation_config import EmulationInfo
from opensn.const.etcd_key import EMU_CONFIG_KEY
from opensn.model.emulation_config import emulation_info_from_json,emulation_info_to_json

def get_emulation_config(etcd_client:Etcd3Client) -> EmulationInfo:
    val,meta = etcd_client.get(EMU_CONFIG_KEY)
    return emulation_info_from_json(val)

def put_emulation_config(etcd_client:Etcd3Client,config: EmulationInfo):
    etcd_client.put(EMU_CONFIG_KEY,emulation_info_to_json(config))