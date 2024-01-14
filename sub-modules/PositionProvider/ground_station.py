from instance import Instance,distance
from const_var import TYPE_GROUND_STATION
from satellite import Satellite
import math

class GroundStation(Instance):

    def __init__(self, instance_id: str, ns: str, node_index: int,latitude: float,longitude: float,altitude:float):
        Instance.__init__(instance_id, TYPE_GROUND_STATION, ns, node_index)
        self.altitude = altitude
        self.longitude = longitude
        self.latitude = latitude
        self.connected_satellite_id = None

    def get_closest_satellite(self,satellites : dict[bytes,Satellite]) -> (Satellite,float):
        shortest_distance = math.inf
        selected_satellite = None
        for (sat_id,satellite) in satellites.items():
            instance_distance = distance(self,satellite)
            if shortest_distance > instance_distance:
                shortest_distance = instance_distance
                selected_satellite = satellite
        return selected_satellite,shortest_distance