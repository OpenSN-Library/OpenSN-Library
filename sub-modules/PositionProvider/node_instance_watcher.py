
import json, time
from satellite import Satellite
from ground_station import GroundStation
from loguru import logger
from threading import Thread,RLock
from link import ISL,GSL,LinkBase,LinksLock,Links
from instance import Instance,InstanceLock,Instances
from satellite import Satellite
from address_allocator import alloc_ipv4,format_ipv4
from instance_config import InstancConfig, EndInfo, LinkInfo
from tools import object2Map

from dependency_client import \
        redis_client,\
        etcd_client

from const_var import \
    NODE_LINK_INFO_KEY_TEMPLATE,\
    NODE_INS_INFO_KEY_TEMPLATE,\
    INS_LINK_ID_FIELD,\
    INS_EXTRA_FIELD,\
    INS_TYPE_FIELD,\
    TYPE_SATELLITE,\
    TYPE_GROUND_STATION,\
    EX_TLE0_KEY,\
    EX_TLE1_KEY,\
    EX_TLE2_KEY,\
    EX_ORBIT_INDEX,\
    EX_SATELLITE_INDEX,\
    LINK_ENDINFO_FIELD,\
    ENDINFO_INSTANCE_TYPE_FIELD,\
    ENDINFO_INSTANCE_ID_FIELD,\
    LINK_PARAMETER_FIELD,\
    NODE_INST_LIST_KEY_TEMPLATE,\
    INS_CONFIG_FIELD,\
    INS_NS_FIELD,\
    EX_ALTITUDE_KEY,\
    EX_LATITUDE_KEY,\
    EX_LONGITUDE_KEY,\
    LINK_CONFIG_FIELD,\
    LINK_CONFIG_ID_FIELD,\
    NODE_INSTANCE_CONFIG_KEY_TEMPLATE,\
    LINK_V4_ADDR_KEY

from const_var import TYPE_SATELLITE

def parse_node_instance_change(key_list : list[str],node_index:int):
    add_key : list[str] = []
    del_key : list[str] = []
    link_id_set : set[list] = set()
    for remote_key in key_list:
        if remote_key not in Instances.keys():
            add_key.append(remote_key)
    for local_key in Instances.keys():
        if local_key not in key_list:
            del_key.append(local_key)
            etcd_client.delete(NODE_INSTANCE_CONFIG_KEY_TEMPLATE%Instances[local_key].node_index+Instances[local_key].instance_id)
    for to_del in del_key:
        del Instances[to_del]
    config_instance_array:list[Instance] = []
    if len(add_key) > 0:
        infos = redis_client.hmget(NODE_INS_INFO_KEY_TEMPLATE%node_index,add_key)
        for index in range(len(add_key)):
            instance_obj_dict = json.loads(infos[index])
            if instance_obj_dict[INS_CONFIG_FIELD][INS_TYPE_FIELD] == TYPE_SATELLITE:
                inst = Satellite(
                    add_key[index],
                    [
                        instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_TLE0_KEY],
                        instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_TLE1_KEY],
                        instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_TLE2_KEY]
                    ],
                    int(instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_ORBIT_INDEX]),
                    int(instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_SATELLITE_INDEX]),
                    instance_obj_dict[INS_NS_FIELD],
                    node_index
                )
                for link_id in instance_obj_dict[INS_CONFIG_FIELD][INS_LINK_ID_FIELD]:
                    link_id_set.add(link_id)
                Instances[add_key[index]]=inst
                config_instance_array.append(inst)
            elif instance_obj_dict[INS_CONFIG_FIELD][INS_TYPE_FIELD] == TYPE_GROUND_STATION:
                inst = GroundStation(
                    add_key[index],
                    instance_obj_dict[INS_NS_FIELD],
                    node_index,
                    float(instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_LATITUDE_KEY]),
                    float(instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_LONGITUDE_KEY]),
                    float(instance_obj_dict[INS_CONFIG_FIELD][INS_EXTRA_FIELD][EX_ALTITUDE_KEY])
                )
                if instance_obj_dict[INS_CONFIG_FIELD][INS_LINK_ID_FIELD] is not None:
                    for link_id in instance_obj_dict[INS_CONFIG_FIELD][INS_LINK_ID_FIELD]:
                        if link_id in Links.keys():
                            link_id_set.add(Links[link_id])
                Instances[add_key[index]]=inst
                config_instance_array.append(inst)

    if len(add_key) > 0:
        for instance_info in config_instance_array:
            logger.info("Update Instance Config of %s"%instance_info.instance_id)
            config = InstancConfig(instance_info.instance_id)
            for link_id,link_info in instance_info.links.items():
                if link_info.config.init_end_infos[0].instance_id == instance_info.instance_id:
                    config.link_infos[link_id] = LinkInfo(link_info.config.address_infos[0][LINK_V4_ADDR_KEY],"")
                    config.end_infos[link_id] = EndInfo(
                        link_info.config.init_end_infos[1].instance_id,
                        link_info.config.init_end_infos[1].instance_type,
                    )
                else:
                    config.link_infos[link_id] = LinkInfo(link_info.config.address_infos[1][LINK_V4_ADDR_KEY],"")
                    config.end_infos[link_id] = EndInfo(
                        link_info.config.init_end_infos[0].instance_id,
                        link_info.config.init_end_infos[0].instance_type,
                    )

            etcd_client.put(
                NODE_INSTANCE_CONFIG_KEY_TEMPLATE%(
                    Instances[instance_info.instance_id].node_index
                )+Instances[instance_info.instance_id].instance_id,
                
                json.dumps(object2Map(config))
            )
                


class NodeInstanceWatcher(Thread):

    def __init__(self,node_index:int):
        Thread.__init__(self)
        self.node_index = node_index
        self.watch_key = NODE_INST_LIST_KEY_TEMPLATE%self.node_index
        self.cancel = None
        self.stop_sig = True

    def terminate(self):
        if self.cancel is not None:
            self.cancel()
            self.cancel = None
        self.stop_sig = True

    def run(self):
        if not self.stop_sig:
            return
        self.stop_sig = False
        val,useless = etcd_client.get(self.watch_key)
        if val is not None:
            node_list = json.loads(val)
            InstanceLock.acquire()
            LinksLock.acquire()
            parse_node_instance_change(node_list,self.node_index)
            LinksLock.release()
            InstanceLock.release()
        while not self.stop_sig:
            try:
                events,cancel = etcd_client.watch(self.watch_key)
                self.cancel = cancel
                for event in events:
                    instance_id_list = json.loads(event.value)
                    InstanceLock.acquire()
                    LinksLock.acquire()
                    try:
                        parse_node_instance_change(instance_id_list,self.node_index)
                    except Exception as e1:
                        logger.error(e1)
                    LinksLock.release()
                    InstanceLock.release()
            except Exception as e:
                logger.error("Watch instance list of node %d Error %s"%(self.node_index,str(e)))
                time.sleep(10)
                cancel()