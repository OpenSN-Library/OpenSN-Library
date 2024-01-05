import os,math

# GEO Config
R_EARTH = 6371000
POLAR_REGION_LATITUDE = 66.5 / 180 * math.pi
LIGHT_SPEED_M_S = 3e8

# Dependency Configs
NODE_INDEX = int(os.getenv("NODE_INDEX"))
ETCD_ADDR = os.getenv("ETCD_ADDR")
ETCD_PORT = int(os.getenv("ETCD_PORT"))
REDIS_ADDR = os.getenv("REDIS_ADDR")
REDIS_PORT = int(os.getenv("REDIS_PORT"))
REDIS_PASSWORD = int(os.getenv("REDIS_PASSWORD"))

# Link Parameter Keys
PARAMETER_KEY_CONNECT = "connect"
PARAMETER_KEY_DELAY = "delay"
PARAMETER_KEY_LOSS = "loss"
PARAMETER_KEY_BANDWIDTH = "bandwidth"

# Link Json Object Fields
LINK_ENDINFO_FIELD = "end_infos"
ENDINFO_INSTANCE_TYPE_FIELD = "instance_type"
ENDINFO_INSTANCE_ID_FIELD = "instance_id"
LINK_PARAMETER_FIELD = "parameter"

# Instance Types
TYPE_SATELLITE = "Satellite"
TYPE_GROUND_STATION = "GroundStation"

# ETCD Keys
NODE_NS_LIST_KEY_TEMPLATE = "/node_%d/ns_list"
NS_POS_KEY_TEMPLATE = "/positions/%s/%s"
NODE_INST_LIST_KEY_TEMPLATE = "/node_%d/instance_list"
NODE_LIST_KEY = "/node_index_list"
NODE_LINK_PARAMETER_KEY_TEMPLATE = "/node_%d/link_paramter"

# Redis Keys
NODE_LINK_INFO_KEY_TEMPLATE= "node_%d_links"
NODE_INS_INFO_KEY_TEMPLATE = "node_%d_instances"

# Instance Json Object Fields
INS_TYPE_FIELD = "type"
INS_NS_FIELD = "namespace"
INS_EXTRA_FIELD = "extra"
EX_TLE0_KEY = "TLE_0"
EX_TLE1_KEY = "TLE_1"
EX_TLE2_KEY = "TLE_2"
EX_ORBIT_INDEX = "OrbitIndex"
EX_SATELLITE_INDEX = "SatelliteIndex"
INS_LINK_ID_FIELD = "link_ids"
INS_CONFIG_FIELD = "config"