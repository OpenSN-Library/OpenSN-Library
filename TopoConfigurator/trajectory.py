
import math
import ephem
import datetime
from opensn.const.const_var import R_EARTH,LIGHT_SPEED_M_S
from opensn.model.position import Position
from instance_types import TYPE_SATELLITE,TYPE_GROUND_STATION
from instance_types import EX_TLE0_KEY,EX_TLE1_KEY,EX_TLE2_KEY,EX_LATITUDE_KEY,EX_LONGITUDE_KEY,EX_ALTITUDE_KEY
from opensn.model.instance import Instance
        
def calculate_postion(instance: Instance,time:datetime.datetime) -> Position:
    ret = Position()
    if instance.type == TYPE_SATELLITE and instance.start:
        ephem_time = ephem.Date(time)
        ephem_obj = ephem.readtle(
            instance.extra[EX_TLE0_KEY],
            instance.extra[EX_TLE1_KEY],
            instance.extra[EX_TLE2_KEY],
        )
        ephem_obj.compute(ephem_time)
        ret.latitude = ephem_obj.sublat
        ret.longitude = ephem_obj.sublong
        ret.altitude = ephem_obj.elevation
    elif instance.type == TYPE_GROUND_STATION:
        ret.latitude = float(instance.extra[EX_LATITUDE_KEY])
        ret.longitude = float(instance.extra[EX_LONGITUDE_KEY])
        ret.altitude = float(instance.extra[EX_ALTITUDE_KEY])
    return ret

def distance_meter(one:Position,another:Position) -> float: # meter
    z1 = (one.altitude+R_EARTH) * math.sin(one.latitude)
    base1 = (one.altitude+R_EARTH) * math.cos(one.latitude)
    x1 = base1 * math.cos(one.longitude)
    y1 = base1 * math.sin(one.longitude)
    z2 = (another.altitude+R_EARTH) * math.sin(another.latitude)
    base2 = (another.altitude+R_EARTH) * math.cos(another.latitude)
    x2 = base2 * math.cos(another.longitude)
    y2 = base2 * math.sin(another.longitude)
    return math.sqrt((x1-x2)**2+(y1-y2)**2+(z1-z2)**2)

def get_propagation_delay_s(distance_meter:float) -> float: # second
    return distance_meter / LIGHT_SPEED_M_S

def select_closest_satellite(
        ground_station:Instance,
        position_map:dict[str,Position],
        instance_map:dict[str,Instance]
    ) -> (str,bool) :
    closet_distance = math.inf
    select_satellite_id = ""
    change = True
    for instance_id,instance_info in instance_map.items():
        if instance_info.type != TYPE_SATELLITE:
            continue
        new_distance = distance_meter(
            position_map[instance_id],
            position_map[ground_station.instance_id],
        )
        if new_distance < closet_distance:
            closet_distance = new_distance
            select_satellite_id = instance_id
    if len(ground_station.connections) < 0 and select_satellite_id == "":
        return "",False
    
    for end_info in ground_station.connections.values():
        if select_satellite_id == end_info.instance_id:
            change = False
    return select_satellite_id,change