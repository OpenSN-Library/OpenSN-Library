"""Standard 控制器（纯 etcd 版）。

这里只有"这个示例自己的策略"：地面站怎么换星、什么时候断链、参数取多少、容器配置长什么样。
平台协议（键、JSON、建链、地址透传、批量提交）由 opensn SDK 负责。
"""

from datetime import datetime
from typing import Optional

from opensn import OpenSN

import container_config
from addressing import AddressPool
from config import BANDWIDTH_BPS, LOSS_RATIO, POLAR_THRESHOLD_DEG, STEP_SECOND
from physics import (
    GROUND_STATION,
    SATELLITE,
    distance_m,
    is_cross_plane_polar,
    nearest_satellite,
    position_of,
    propagation_delay_s,
)


def setup(sn: OpenSN) -> int:
    sn.wait_ready()
    sn.emulation.wait_running(True, timeout=60)
    pushed = 0
    for inst in sn.instances.started():
        if not inst.config_exists():
            inst.push_config(container_config.build(sn, inst))
            pushed += 1
    print(f"平台就绪：{len(sn.nodes.list())} 台机器，{len(sn.instances.started())} 个已启动实例，补配置 {pushed} 份")
    return pushed


def switch_station(sn: OpenSN, station, target, pool: AddressPool):
    """换星策略：复用旧链路的地址；没有旧链路就新分配一对。"""
    old = None
    links = sn.links.of(station)
    if links:
        old = links[0]
    old_peer = None
    if old is not None:
        mine = 0 if old.end_infos[0].instance_id == station.instance_id else 1
        address_infos = [dict(old.address_infos[mine]), dict(old.address_infos[1 - mine])]
        old_peer = sn.instances.get(old.end_infos[1 - mine].instance_id)
        old.remove()
    else:
        first, second = pool.alloc()
        address_infos = [{"IPV4": first}, {"IPV4": second}]
    link = sn.links.create(station, target, address_infos=address_infos)
    # 连接关系变了：给站、新卫星、旧卫星各下发一次配置
    for inst in (station, target, old_peer):
        if inst is not None:
            inst.push_config(container_config.build(sn, inst))
    return link


def step_once(sn: OpenSN, step, pool: AddressPool) -> None:
    positions = {}
    for inst in sn.instances.started():
        positions[inst.instance_id] = position_of(inst, datetime.now())
        step.positions.set(inst, positions[inst.instance_id])

    satellites = [inst for inst in sn.instances.of_type(SATELLITE) if inst.start]
    for station in sn.instances.of_type(GROUND_STATION):
        if not station.start:
            continue
        target = nearest_satellite(station, satellites, positions)
        if target is not None and not station.connected_to(target):
            switch_station(sn, station, target, pool)

    for link in sn.links.active():
        a, b = link.endpoints
        if a.instance_id not in positions or b.instance_id not in positions:
            continue
        delay_s = propagation_delay_s(distance_m(positions[a.instance_id], positions[b.instance_id]))
        link.set(
            connect=not is_cross_plane_polar(link, positions, POLAR_THRESHOLD_DEG),
            delay_s=delay_s,
            bandwidth_bps=BANDWIDTH_BPS,
            loss_ratio=LOSS_RATIO,
        )


def main(sn: Optional[OpenSN] = None) -> None:
    client = sn or OpenSN.from_env()
    setup(client)
    pool = AddressPool(client)
    try:
        for step in client.steps(interval=STEP_SECOND):
            step_once(client, step, pool)
    finally:
        if sn is None:
            client.close()


if __name__ == "__main__":
    main()
