
import json, time
from satellite import Satellite
from ground_station import GroundStation
from loguru import logger
from threading import Thread,RLock
from link import ISL,GSL,LinkBase,Links,LinksLock
from instance import InstanceLock,Instances
from address_allocator import alloc_ipv4,format_ipv4
from instance_config import InstancConfig, EndInfo, LinkInfo
from tools import object2Map

from dependency_client import \
        redis_client,\
        etcd_client

from const_var import \
    NODE_LINK_INFO_KEY_TEMPLATE,\
    TYPE_SATELLITE,\
    LINK_ENDINFO_FIELD,\
    ENDINFO_INSTANCE_TYPE_FIELD,\
    NODE_LINK_LIST_KEY_TEMPLATE,\
    LINK_CONFIG_FIELD,\
    NODE_INSTANCE_CONFIG_KEY_TEMPLATE,\
    LINK_V4_ADDR_KEY

from const_var import TYPE_SATELLITE

def create_link_from_json(json_seq: bytes) -> LinkBase:
    link_dict:dict = json.loads(json_seq)
    if link_dict[LINK_CONFIG_FIELD][LINK_ENDINFO_FIELD][0][ENDINFO_INSTANCE_TYPE_FIELD] == TYPE_SATELLITE and \
        link_dict[LINK_CONFIG_FIELD][LINK_ENDINFO_FIELD][1][ENDINFO_INSTANCE_TYPE_FIELD] == TYPE_SATELLITE :
        ret = ISL(link_dict)
        ret.is_inter_orbit = ret.config.init_end_infos[0].instance_id != ret.config.init_end_infos[0].instance_id
    else:
        ret = GSL(link_dict)
    return ret

def parse_node_link_change(key_list : list[str],node_index:int):
    add_key : list[str] = []
    del_key : list[str] = []
    dirty_instance_key:set[str] = set()

    for remote_key in key_list:
        if remote_key not in Links.keys():
            add_key.append(remote_key)
    for local_key in Links.keys():
        if local_key not in key_list:
            link_info = Links[local_key]
            del_key.append(local_key)
            for end_info in link_info.config.init_end_infos:
                if end_info.instance_id in Instances.keys():
                    dirty_instance_key.add(end_info.instance_id)
    

    for key in del_key:
        del Links[key]
        
    
    if len(add_key) > 0:
        link_info_seqs = redis_client.hmget(NODE_LINK_INFO_KEY_TEMPLATE%node_index,add_key)
        for link_info_index in range(len(link_info_seqs)):
            if link_info_seqs[link_info_index] is None:
                continue
            subnet = alloc_ipv4(30)
            link_obj = create_link_from_json(link_info_seqs[link_info_index])
            link_obj.config.address_infos[0] = {LINK_V4_ADDR_KEY:format_ipv4(subnet[1],30)}
            link_obj.config.address_infos[1] = {LINK_V4_ADDR_KEY:format_ipv4(subnet[2],30)}
            Links[add_key[link_info_index]] = link_obj
            if link_obj.config.init_end_infos[0].instance_id in Instances.keys():
                Instances[link_obj.config.init_end_infos[0].instance_id].links[add_key[link_info_index]] = link_obj
                dirty_instance_key.add(link_obj.config.init_end_infos[0].instance_id)
            if link_obj.config.init_end_infos[1].instance_id in Instances.keys():
                Instances[link_obj.config.init_end_infos[1].instance_id].links[add_key[link_info_index]] = link_obj
                dirty_instance_key.add(link_obj.config.init_end_infos[1].instance_id)
            redis_client.hset(NODE_LINK_INFO_KEY_TEMPLATE%node_index,add_key[link_info_index],json.dumps(object2Map(link_obj)))

    if len(add_key) > 0:
        for instance_id in dirty_instance_key:
            instance_info = Instances[instance_id]
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
                


class NodeLinkWatcher(Thread):

    def __init__(self,node_index:int):
        Thread.__init__(self)
        self.node_index = node_index
        self.watch_key = NODE_LINK_LIST_KEY_TEMPLATE%self.node_index
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
            parse_node_link_change(node_list,self.node_index)
            LinksLock.release()
            InstanceLock.release()
        while not self.stop_sig:
            try:
                events,cancel = etcd_client.watch(self.watch_key)
                self.cancel = cancel
                for event in events:
                    link_id_list = json.loads(event.value)
                    InstanceLock.acquire()
                    LinksLock.acquire()
                    try:
                        parse_node_link_change(link_id_list,self.node_index)
                    except Exception as e1:
                        logger.error(e1)
                    LinksLock.release()
                    InstanceLock.release()
            except Exception as e:
                logger.error("Watch instance list of node %d Error %s"%(self.node_index,str(e)))
                time.sleep(10)
                cancel()