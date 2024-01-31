import requests,json
import etcd3
from concurrent.futures import ThreadPoolExecutor
from satellite_emulator.const.link_type import VLINK_TYPE
from satellite_emulator.model.node import Node
from satellite_emulator.model.instance import Instance,InstanceRuntime,ConnectionInfo
from satellite_emulator.model.link import LinkBase,create_new_link
from satellite_emulator.model.emulation_config import EmulationInfo
from satellite_emulator.model.position import Position
from satellite_emulator.synchronizer.sync_instance import get_instance,get_instance_map,put_instance
from satellite_emulator.synchronizer.sync_instance import get_instance_runtime,get_instance_runtime_map
from satellite_emulator.synchronizer.sync_instance import put_instance_config,put_instance_config_if_not_exist
from satellite_emulator.synchronizer.sync_position import get_position,get_position_map
from satellite_emulator.synchronizer.sync_node import get_node_map,get_node
from satellite_emulator.synchronizer.sync_link import get_link,get_link_map,put_link,put_link_parameter,remove_link
from satellite_emulator.synchronizer.sync_config import get_emulation_config,put_emulation_config

class EmulatorOperator:

    def __init__(self,addr,port,async_pool_size = 128) -> None:
        response = requests.get("http://%s:%s/api/platform/status"%(addr,port))
        if response.status_code != 200 :
            raise RuntimeError("emulation platform is not ready, resp code is %d"%response.status_code)
        response = requests.get("http://%s:%s/api/platform/address/etcd"%(addr,port))
        if response.status_code != 200 :
            raise RuntimeError("get etcd address info error, resp code is %d"%response.status_code)
        etcd_info = json.loads(response.content)
        self.etcd_client = etcd3.client(host=etcd_info["data"]["address"],port=etcd_info["data"]["port"])
        self.__pool = ThreadPoolExecutor(max_workers=async_pool_size)


    def close(self):
        self.etcd_client.close()

    def get_emulation_config(self) -> EmulationInfo:
        return get_emulation_config(self.etcd_client)

    def put_emualtion_config(self,config: EmulationInfo):
        return put_emulation_config(self.etcd_client,config)

    def put_emualtion_config_async(self,config: EmulationInfo):
        return self.__pool.submit(put_emulation_config,self.etcd_client,config)

    def get_node_map(self) -> dict[str,Node]:
        return get_node_map(self.etcd_client)
    
    def get_node(self,node_index: int) -> Node:
        return get_node(self.etcd_client,node_index)
    
    def get_instance(self,node_index:int,instance_id:str) -> Instance:
        return get_instance(self.etcd_client,node_index,instance_id)
    
    def get_instance_map(self,node_index:int) -> dict[str,Instance]:
        return get_instance_map(self.etcd_client,node_index)
    
    def get_link_map(self,node_index:int) -> dict[str,LinkBase]:
        return get_link_map(self.etcd_client,node_index)
    
    def get_link(self,node_index:int,link_id:str) -> LinkBase:
        return get_link(self.etcd_client,node_index,link_id)
    
    def put_link(self,link_base: LinkBase):
        return put_link(self.etcd_client,link_base)
        
    def put_link_async(self,link_base: LinkBase):
        return self.__pool.submit(put_link,self.etcd_client,link_base)

    def put_link_parameter(self,node_index:int,link_id:str,parameter: dict[str,int]):
        return put_link_parameter(self.etcd_client,node_index,link_id,parameter)

    def put_link_parameter_async(self,node_index:int,link_id:str,parameter: dict[str,int]):
        return self.__pool.submit(put_link_parameter,self.etcd_client,node_index,link_id,parameter)

    def put_instance(self,instance: Instance):
        return put_instance(self.etcd_client,instance)

    def put_instance_async(self,instance:Instance):
        return self.__pool.submit(put_instance,self.etcd_client,instance)

    def get_instance_runtime(self, node_index:int, instance_id: str):
        return get_instance_runtime(self.etcd_client,node_index,instance_id)

    def get_instance_runtime_map(self,node_index: int):
        return get_instance_runtime_map(self.etcd_client,node_index)

    def get_position_map(self) -> dict[str,Position]:
        return get_position_map(self.etcd_client)

    def get_position(self, instance_id: str) -> Position:
        return get_position(self.etcd_client,instance_id)
    
    def put_instance_config(self,node_index:int,instance_id:str,config_seq:str):
        return put_instance_config(self.etcd_client,node_index,instance_id,config_seq)

    def put_instance_config_if_not_exist(self,node_index:int,instance_id:str,config_seq:str):
        return put_instance_config_if_not_exist(self.etcd_client,node_index,instance_id,config_seq)

    def put_instance_config_async(self,node_index:int,instance_id:str,config_seq:str):
        return self.__pool.submit(put_emulation_config,node_index,instance_id,config_seq)
    def enable_link_between(
            self,
            node_index1:int,
            instance_id1:str,
            node_index2,
            instance_id2:str,
            address_info1 = {},
            address_info2 = {},
            link_type = VLINK_TYPE,
            init_parameter:dict[str,int] = {}):
        instance1 = get_instance(self.etcd_client,node_index1,instance_id1)
        instance2 = get_instance(self.etcd_client,node_index2,instance_id2)
        new_link_array = create_new_link(
            instance1.node_index,
            instance1.instance_id,
            instance1.type,
            instance2.node_index,
            instance2.instance_id,
            instance2.type,
            link_type=link_type,
            address_info1=address_info1,
            address_info2=address_info2,
            init_parameter=init_parameter
        )
        for link_info in new_link_array:
            put_link(self.etcd_client,link_info)
        connect_info_1 = ConnectionInfo()
        connect_info_1.end_node_index = instance2.node_index
        connect_info_1.instance_id = instance2.instance_id
        connect_info_1.instance_type = instance2.type
        connect_info_1.link_id = new_link_array[0].link_id
        instance1.connections[new_link_array[0].link_id] = connect_info_1
        put_instance(self.etcd_client,instance1)
        connect_info_2 = ConnectionInfo()
        connect_info_2.end_node_index = instance1.node_index
        connect_info_2.instance_id = instance1.instance_id
        connect_info_2.instance_type = instance1.type
        connect_info_2.link_id = new_link_array[0].link_id
        instance2.connections[new_link_array[0].link_id] = connect_info_1
        put_instance(self.etcd_client,instance2)
            

    def disable_link_between(self,node_index1:int,instance_id1:str,node_index2,instance_id2:str) -> dict[str,LinkBase]:
        instance1 = get_instance(self.etcd_client,node_index1,instance_id1)
        instance2 = get_instance(self.etcd_client,node_index2,instance_id2)
        ret:dict[str,LinkBase] = {}
        delete_list: list[str] = []
        for link_id in instance1.connections.keys():
            if link_id in instance2.connections:
                delete_list.append(link_id)
                ret[link_id] = get_link(self.etcd_client,instance1.node_index,link_id)
        for del_id in delete_list:
            remove_link(self.etcd_client,node_index1,link_id)
            if instance1.node_index != instance2.node_index:
                remove_link(self.etcd_client,node_index2,link_id)
            del instance1.connections[link_id]
            del instance2.connections[link_id]
        put_instance(self.etcd_client,instance1)
        put_instance(self.etcd_client,instance2)
        return ret
