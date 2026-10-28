"""用内存 etcd + ephem 桩跑一遍改写后的 Standard 主循环。"""

import json
import sys
import types
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
sys.path.insert(0, str(ROOT / "Standard"))
sys.path.insert(0, str(ROOT / "opensn" / "tests"))

from fake_etcd import FakeEtcd  # noqa: E402

from opensn import OpenSN  # noqa: E402
from opensn.schema.configs import InstanceSpec, LinkSpec  # noqa: E402


class _Body:
    def compute(self, when):
        self.sublat = 0.1
        self.sublong = 0.2
        self.elevation = 550000.0


EPHEM = types.SimpleNamespace(Date=lambda value: value, readtle=lambda *tle: _Body())
sys.modules.setdefault("ephem", EPHEM)

import main as StandardMain  # noqa: E402


def node(index, free=10):
    return {
        "node_index": index,
        "free_instance": free,
        "is_master_node": index == 0,
        "l_3_addr_v_4": "",
        "l_3_addr_v_6": "",
        "l_2_addr": "",
        "ns_instance_map": {},
        "ns_link_map": {},
        "node_link_device_info": {},
        "cpu_core": 8,
    }


EMU = {
    "running": True,
    "type_config": {
        "Satellite": {"image": "sat:1", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
        "GroundStation": {"image": "gs:1", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
    },
}


class TestStandardSmoke(unittest.TestCase):
    def setUp(self):
        self.fake = FakeEtcd()
        self.fake.seed("/nodes/0", json.dumps(node(0)))
        self.fake.seed("/emulation_config", json.dumps(EMU))
        self.sn = OpenSN(client=self.fake, watch=False, poll_interval=60)
        sat = InstanceSpec("Satellite", extra={"TLE_0": "A", "TLE_1": "B", "TLE_2": "C", "OrbitIndex": "0", "Area": "1.1.1.1"})
        gs = InstanceSpec("GroundStation", extra={"latitude": "10", "longitude": "20", "altitude": "100"})
        self.sn.emulation.deploy_topology(
            [sat, gs],
            [LinkSpec(0, 1, address_infos=[{"IPV4": "10.0.0.1/30"}, {"IPV4": "10.0.0.2/30"}])],
        )

    def tearDown(self):
        self.sn.close()

    def test_one_step(self):
        StandardMain.setup(self.sn)
        pool = StandardMain.AddressPool(self.sn)
        steps = self.sn.steps(interval=0.05)
        step = next(steps)
        StandardMain.step_once(self.sn, step, pool)
        next(steps)  # 提交写
        steps.close()

        positions = self.fake.dump("/position")
        self.assertTrue(positions, "位置应该被写入")
        tables = [v for k, v in self.fake.dump("/node_0").items() if "/link_parameter/" in k]
        self.assertTrue(tables, "链路参数应该被写入")
        table = json.loads(tables[0])
        self.assertIn("delay", table)
        self.assertIn("connect", table)
        configs = self.fake.dump("/instance_config")
        self.assertTrue(configs, "容器配置应该被下发")


if __name__ == "__main__":
    unittest.main()
