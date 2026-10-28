"""本示例自己的物理计算：TLE -> 位置、距离、时延、最近卫星、极区判定。

这是**用户代码**，不是平台协议：换一套轨道模型、换一种选星策略，只改这里。
"""

import math
from typing import Dict, List, Optional

from opensn.schema.entities import Instance, Position

SATELLITE = "Satellite"
GROUND_STATION = "GroundStation"

TLE_0 = "TLE_0"
TLE_1 = "TLE_1"
TLE_2 = "TLE_2"
ORBIT_INDEX = "OrbitIndex"
AREA = "Area"
LATITUDE = "latitude"
LONGITUDE = "longitude"
ALTITUDE = "altitude"

R_EARTH_M = 6371000.0
LIGHT_SPEED_M_S = 3e8


def ground_position(instance: Instance) -> Position:
    return Position.from_degrees(
        float(instance.extra.get(LATITUDE, 0.0)),
        float(instance.extra.get(LONGITUDE, 0.0)),
        float(instance.extra.get(ALTITUDE, 0.0)),  # 米
    )


def satellite_position(instance: Instance, when) -> Position:
    import ephem  # 示例用 SGP4；换成自己的轨道模型也可以

    tle = (instance.extra.get(TLE_0), instance.extra.get(TLE_1), instance.extra.get(TLE_2))
    if not all(tle):
        raise ValueError(f"instance {instance.instance_id} has no TLE in extra")
    body = ephem.readtle(*tle)
    body.compute(ephem.Date(when))
    return Position(latitude=float(body.sublat), longitude=float(body.sublong), altitude=float(body.elevation))


def position_of(instance: Instance, when) -> Position:
    if instance.type == SATELLITE:
        return satellite_position(instance, when)
    return ground_position(instance)


def distance_m(one: Position, another: Position) -> float:
    z1 = (one.altitude + R_EARTH_M) * math.sin(one.latitude)
    base1 = (one.altitude + R_EARTH_M) * math.cos(one.latitude)
    x1 = base1 * math.cos(one.longitude)
    y1 = base1 * math.sin(one.longitude)
    z2 = (another.altitude + R_EARTH_M) * math.sin(another.latitude)
    base2 = (another.altitude + R_EARTH_M) * math.cos(another.latitude)
    x2 = base2 * math.cos(another.longitude)
    y2 = base2 * math.sin(another.longitude)
    return math.sqrt((x1 - x2) ** 2 + (y1 - y2) ** 2 + (z1 - z2) ** 2)


def propagation_delay_s(distance: float) -> float:
    return float(distance) / LIGHT_SPEED_M_S


def nearest_satellite(ground_station: Instance, satellites: List[Instance], positions: Dict[str, Position]) -> Optional[Instance]:
    best = None
    best_distance = math.inf
    gs_position = positions.get(ground_station.instance_id)
    if gs_position is None:
        return None
    for sat in satellites:
        sat_position = positions.get(sat.instance_id)
        if sat_position is None:
            continue
        distance = distance_m(sat_position, gs_position)
        if distance < best_distance:
            best_distance = distance
            best = sat
    return best


def is_cross_plane_polar(link, positions: Dict[str, Position], threshold_deg: float = 66.5) -> bool:
    """跨轨道面、且任一端进入极区 -> 视为要断的链路。这是示例自己的规则。"""
    threshold = math.radians(float(threshold_deg))
    a, b = link.endpoints
    if a.type != SATELLITE or b.type != SATELLITE:
        return False
    if a.extra.get(ORBIT_INDEX) == b.extra.get(ORBIT_INDEX):
        return False
    for inst in (a, b):
        position = positions.get(inst.instance_id)
        if position is not None and abs(position.latitude) > threshold:
            return True
    return False
