"""示例自己的"实例类型约定"。

这些不是平台协议：类型名、extra 字段名都是本示例与容器镜像之间的约定。
SDK 只负责把它们原样存进 etcd。
"""

TYPE_SATELLITE = "Satellite"
TYPE_GROUND_STATION = "GroundStation"
TYPE_GROUND_TERMINAL = "GroundTerminal"

# 卫星 extra
EX_TLE0_KEY = "TLE_0"
EX_TLE1_KEY = "TLE_1"
EX_TLE2_KEY = "TLE_2"
EX_ORBIT_INDEX = "OrbitIndex"
EX_SATELLITE_INDEX = "SatelliteIndex"
EX_AREA_KEY = "Area"

# 地面站 extra
EX_LATITUDE_KEY = "latitude"
EX_LONGITUDE_KEY = "longitude"
EX_ALTITUDE_KEY = "altitude"
