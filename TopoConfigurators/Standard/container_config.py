"""本示例的容器配置格式。

这是**用户与镜像之间的约定**（container-base/satellite-router 与 ground-station 认这个结构），
不是平台协议。平台只负责把这里生成的内容原样写到容器内 /share/topo/<instance_id>。
换镜像/换配置格式，只改这里。
"""

from typing import Dict

from opensn import OpenSN

AREA = "Area"


def build(sn: OpenSN, inst) -> dict:
    config = {"instance_id": inst.instance_id, "link_infos": {}, "end_infos": {}}
    for link in sn.links.of(inst):
        mine = None
        for i, end in enumerate(link.end_infos):
            if end.instance_id == inst.instance_id:
                mine = i
        if mine is None:
            continue
        peer_end = link.end_infos[1 - mine]
        address = (link.address_infos[mine] or {}).get("IPV4")
        if address:
            config["link_infos"][link.link_id] = {"IPV4": address}
        peer_info = {"instance_id": peer_end.instance_id, "type": peer_end.instance_type}
        try:
            peer = sn.instances.get(peer_end.instance_id)
        except Exception:
            peer = None
        if peer is not None and AREA in peer.extra:
            peer_info["area"] = peer.extra[AREA]
        config["end_infos"][link.link_id] = peer_info
    if AREA in inst.extra:
        config["area"] = inst.extra[AREA]
    return config
