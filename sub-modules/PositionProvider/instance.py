from link import LinkBase
from const_var import R_EARTH,LIGHT_SPEED_M_S
import math
from threading import RLock
class Instance:

    def __init__(self,instance_id:str,instance_type:str,ns:str,node_index:int,):
        self.instance_id = instance_id
        self.type = instance_type
        self.node_index = node_index
        self.latitude = 0.0 # radius
        self.longitude = 0.0 # radius
        self.altitude = 0.0 # meter
        self.links: dict[str,LinkBase] = {}
        self.namespace = ns

    def get_position_dict(self) -> dict[str,float]:
        return {
            "latitude" : self.latitude,
            "longitude" : self.longitude,
            "altitude" : self.altitude
        }

def distance(one:Instance,another:Instance) -> float: # meter
    z1 = (one.altitude+R_EARTH) * math.sin(one.latitude)
    base1 = (one.altitude+R_EARTH) * math.cos(one.latitude)
    x1 = base1 * math.cos(one.longitude)
    y1 = base1 * math.sin(one.longitude)
    z2 = (another.altitude+R_EARTH) * math.sin(another.latitude)
    base2 = (another.altitude+R_EARTH) * math.cos(another.latitude)
    x2 = base2 * math.cos(another.longitude)
    y2 = base2 * math.sin(another.longitude)
    return math.sqrt((x1-x2)**2+(y1-y2)**2+(z1-z2)**2)

def get_propagation_delay(distance_meter:float) -> float: # second
    return distance_meter / LIGHT_SPEED_M_S

InstanceLock = RLock()
Instances: dict[bytes,Instance] = {}
