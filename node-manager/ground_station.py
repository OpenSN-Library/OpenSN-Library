from docker_client import DockerClient
from satellite_node import SatelliteNode
from subnet_allocator import ip2str
from const_var import LATITUDE_KEY,LONGITUDE_KEY,HEIGHT_KEY,R_EARTH
import json
import math
import time
from loguru import logger
ground_stations = []

class GroundStation:
    DockerCli : DockerClient = None
    GroundStationCounter : int = 0
    def __init__(self,node_id: str, long: float,lat: float, cont_id: str, net_id: str) -> None:
        self.latitude = lat / 180 * math.pi
        self.longitude = long / 180 * math.pi
        self.node_id = node_id
        self.container_id = cont_id
        self.network_id = net_id
        self.connected_satellite_id = None
        self.connected_node_id = None
    
    def disconnect_satellite(self):
        if self.connected_satellite_id is None or GroundStation.DockerCli is None :
            return
        GroundStation.DockerCli.disconnect_node(self.connected_satellite_id,self.network_id)
        self.connected_satellite_id = None

    def connect_satellite(self,sat_node_id:str, sat_cont_id: str):
        if GroundStation.DockerCli is None:
            return
        GroundStation.DockerCli.connect_node(sat_cont_id,self.network_id,"sat")
        self.connected_satellite_id = sat_cont_id
        self.connected_node_id = sat_node_id

    def switch_satellite(self, sat_node_id:str,sat_cont_id: str):
        logger.info("%s switch from %s to %s"%(self.node_id,self.connected_node_id,sat_node_id))
        if sat_cont_id == self.connected_satellite_id:
            return
        old_sat_id = self.connected_satellite_id
        self.disconnect_satellite()
        self.connect_satellite(sat_node_id,sat_cont_id)

    
def distance(position_data: dict, ground: GroundStation) -> float:
    z1 = (position_data[HEIGHT_KEY]+R_EARTH) * math.sin(position_data[LATITUDE_KEY])
    base1 = (position_data[HEIGHT_KEY]+R_EARTH) * math.cos(position_data[LATITUDE_KEY])
    x1 = base1 * math.cos(position_data[LONGITUDE_KEY])
    y1 = base1 * math.sin(position_data[LONGITUDE_KEY])
    z2 = R_EARTH * math.sin(ground.latitude)
    base2 = R_EARTH * math.cos(ground.latitude)
    x2 = base2 * math.cos(ground.longitude)
    y2 = base2 * math.sin(ground.longitude)
    return math.sqrt((x1-x2)**2+(y1-y2)**2+(z1-z2)**2)

def ground_select(satellites: list,position_data: dict,grounds: list):
    # logger.info("Enter ground daemon.")
    connections = {}
    for ground in grounds:
        if ground.connected_satellite_id is None:
            shortest_id = None
            shortest_dis = math.inf
            shortest_cont = None
        else:
            shortest_id = ground.connected_node_id
            shortest_cont = ground.connected_satellite_id
            shortest_dis = distance(position_data[shortest_id],ground)
        for sat in satellites:
            new_distance = distance(position_data[sat.node_id],ground)
            # logger.info("%s new %f old: %f"%(sat.node_id,new_distance,shortest_dis))
            if new_distance < shortest_dis :
                
                shortest_id = sat.node_id
                shortest_cont = sat.container_id
                shortest_dis = new_distance
        ground.switch_satellite(shortest_id,shortest_cont) 
        connections[ground.node_id] = shortest_id
    return connections

def create_station_from_json(docker_client : DockerClient, path: str):
    grounds = {}
    json_config = open(path,"r")
    config_bytes = json_config.read()
    ground_list = json.loads(config_bytes)
    for ground in ground_list:
        new_ground = create_ground_station(docker_client,ground[LONGITUDE_KEY],ground[LATITUDE_KEY])
        grounds[new_ground.node_id] = {
            "lat": new_ground.latitude,
            "long": new_ground.longitude
        }
    return grounds

def create_ground_station(docker_client : DockerClient, long: float, lat: float):
    if GroundStation.DockerCli is None :
        GroundStation.DockerCli = docker_client
    ground_id = "ground_" + str(GroundStation.GroundStationCounter)
    GroundStation.GroundStationCounter += 1
    container_id = docker_client.create_ground_container(ground_id)
    net_id, net_ip = docker_client.create_network(ground_id)
    docker_client.connect_node(container_id,net_id,"conn")
    new_station = GroundStation(ground_id, long, lat, container_id, net_id)
    ground_stations.append(new_station)
    docker_client.exec_cmd(container_id=container_id,cmd=["route","add","default","gw",ip2str(net_ip+2)])
    return new_station