from collections.abc import Callable, Iterable, Mapping
import json, time
from typing import Any
from instance import Instance
from satellite import Satellite
from loguru import logger
from threading import Thread,RLock
from link import ISL,GSL,Link

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
    INS_NS_FIELD

from const_var import TYPE_SATELLITE


MovingInstancesLock = RLock()
MovingInstances: dict[bytes,Instance] = {}


'''
type Instance struct {
	Config      InstanceConfig `json:"config"`
	ContainerID string         `json:"container_id"`
	Pid         int            `json:"pid"`
	NodeID      uint32         `json:"node_id"`
	Namespace   string         `json:"namespace"`
	LinksID     []string       `json:"links_id"`
}

type InstanceConfig struct {
	InstanceID         string                       `json:"instance_id"`
	Name               string                       `json:"name"`
	Type               string                       `json:"type"`
	Image              string                       `json:"image"`
	PositionChangeable bool                         `json:"position_changeable"`
	Extra              map[string]string            `json:"extra"`
	LinkIDs            []string                     `json:"link_ids"`
	DeviceInfo         map[string]DeviceRequireInfo `json:"device_need"`
}

Type = SATELLITE = "Satellite"

Extra Has
TLE_0 -> TLE_LINE0
TLE_1 -> TLE_LINE1
TLE_2 -> TLE_LINE2
OrbitIndex -> Oribit Index Of Satellite
SatelliteIndex -> Index Of Satellite in its Orbit
'''

def create_link_from_json(json_seq: bytes) -> Link:
    link_dict = json.loads(json_seq)
    if link_dict[LINK_ENDINFO_FIELD][0][ENDINFO_INSTANCE_TYPE_FIELD] == TYPE_SATELLITE and \
        link_dict[LINK_ENDINFO_FIELD][1][ENDINFO_INSTANCE_TYPE_FIELD] == TYPE_SATELLITE :
        ret = ISL(
            [
                link_dict[LINK_ENDINFO_FIELD][0][ENDINFO_INSTANCE_ID_FIELD],
                link_dict[LINK_ENDINFO_FIELD][1][ENDINFO_INSTANCE_ID_FIELD]
            ],
            link_dict[LINK_PARAMETER_FIELD],
            False
        )
        if MovingInstances[ret.instance_id[0]].orbit_index != MovingInstances[ret.instance_id[1]].orbit_index :
            ret.is_inter_orbit = True
        else:
            ret.is_inter_orbit = False
    else:
        ret = GSL(
            [
                link_dict[LINK_ENDINFO_FIELD][0][ENDINFO_INSTANCE_ID_FIELD],
                link_dict[LINK_ENDINFO_FIELD][1][ENDINFO_INSTANCE_ID_FIELD]
            ],
            link_dict[LINK_PARAMETER_FIELD],
            True
        )
        if ret.instance_id[0] == "" or ret.instance_id[1] == "" :
            ret.is_float = True
        else:
            ret.is_float = False
    return ret

def parse_node_instance_change(key_list : list[str],node_index:int):
    add_key : list[str] = []
    del_key : list[str] = []
    link_id_set : set[list] = set()
    for remote_key in key_list:
        if remote_key not in MovingInstances.keys():
            add_key.append(remote_key)
    for local_key in MovingInstances.keys():
        if local_key not in key_list:
            del_key.append(local_key)
    for to_del in del_key:
        del MovingInstances[to_del]
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
                MovingInstances[add_key[index]]=inst
            elif instance_obj_dict[INS_CONFIG_FIELD][INS_TYPE_FIELD] == TYPE_GROUND_STATION:
                pass
    if len(link_id_set) > 0:
        link_id_list = list(link_id_set)
        link_info_seqs = redis_client.hmget(NODE_LINK_INFO_KEY_TEMPLATE%node_index,link_id_list)
        for link_info_index in range(len(link_info_seqs)):
            if link_info_seqs[link_info_index] is None:
                continue
            link_obj = create_link_from_json(link_info_seqs[link_info_index])
            MovingInstances[link_obj.instance_id[0]].links[link_id_list[link_info_index]] = link_obj
            MovingInstances[link_obj.instance_id[1]].links[link_id_list[link_info_index]] = link_obj
        


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
            MovingInstancesLock.acquire()
            parse_node_instance_change(node_list,self.node_index)
            MovingInstancesLock.release()
        while not self.stop_sig:
            try:
                events,cancel = etcd_client.watch(self.watch_key)
                self.cancel = cancel
                for event in events:
                    instance_id_list = json.loads(event.value)
                    parse_node_instance_change(instance_id_list,self.node_index)
                    MovingInstancesLock.acquire()
                    MovingInstancesLock.release()
                    print(MovingInstances)
            except Exception as e:
                logger.error("Watch instance list of node %d Error %s"%(self.node_index,str(e)))
                time.sleep(10)
                cancel()