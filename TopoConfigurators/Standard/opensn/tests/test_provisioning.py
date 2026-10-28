import json
import unittest

from opensn import OpenSN
from opensn.errors import ValidationError
from opensn.schema.configs import InstanceSpec, LinkSpec
from opensn.tests.fake_etcd import FakeEtcd


def node_payload(node_index, free, cpu=8):
    return {
        "node_index": node_index,
        "free_instance": free,
        "is_master_node": node_index == 0,
        "l_3_addr_v_4": "",
        "l_3_addr_v_6": "",
        "l_2_addr": "",
        "ns_instance_map": {},
        "ns_link_map": {},
        "node_link_device_info": {"fplink": 2},
        "cpu_core": cpu,
    }


EMU = {
    "running": False,
    "type_config": {
        "TypeA": {"image": "sat:latest", "container_envs": {}, "resource_limit": {"nano_cpu": 50000000, "memory_byte": 134217728}},
        "TypeB": {"image": "gs:latest", "container_envs": {}, "resource_limit": {"nano_cpu": 50000000, "memory_byte": 134217728}},
    },
}


class TestProvisioning(unittest.TestCase):
    def setUp(self):
        self.fake = FakeEtcd()
        self.fake.seed("/nodes/0", json.dumps(node_payload(0, 3)))
        self.fake.seed("/nodes/1", json.dumps(node_payload(1, 10)))
        self.fake.seed("/emulation_config", json.dumps(EMU))
        self.sn = OpenSN(client=self.fake, watch=False, poll_interval=60)

    def tearDown(self):
        self.sn.close()

    def _deploy(self, dry_run=False):
        instances = [InstanceSpec("TypeA", extra={"opaque_field": str(i)}) for i in range(3)]
        instances += [InstanceSpec("TypeB", extra={"user_lat": "0", "longitude": "0"}) for _ in range(2)]
        links = [LinkSpec(0, 1), LinkSpec(2, 3)]
        return self.sn.emulation.deploy_topology(instances, links, dry_run=dry_run)

    def test_deploy_places_instances_and_links(self):
        result = self._deploy()
        self.assertEqual(len(result.instance_ids), 5)
        self.assertEqual(len(result.link_ids), 2)
        inst_raw = self.fake.dump("/node_")
        instances = {k: json.loads(v) for k, v in inst_raw.items() if "/instances/" in k}
        links = {k: json.loads(v) for k, v in inst_raw.items() if "/links/" in k}
        self.assertEqual(len(instances), 5)
        # 节点 0 只有 3 个容量，前 3 个实例都在这；后 2 个去节点 1
        placed = sorted(v["node_index"] for v in instances.values())
        self.assertEqual(placed, [0, 0, 0, 1, 1])
        # 容量扣减，其它字段保留
        nodes = {k: json.loads(v) for k, v in self.fake.dump("/nodes").items()}
        self.assertEqual(nodes["/nodes/0"]["free_instance"], 0)
        self.assertEqual(nodes["/nodes/1"]["free_instance"], 8)
        self.assertEqual(nodes["/nodes/0"]["cpu_core"], 8)
        self.assertEqual(nodes["/nodes/0"]["node_link_device_info"], {"fplink": 2})
        # 链路：第 2 条跨机 -> 两条记录，link_index 相同
        by_id = {}
        for key, rec in links.items():
            by_id.setdefault(rec["link_id"], []).append(rec)
        self.assertEqual(len(by_id), 2)
        cross = [records for records in by_id.values() if len(records) == 2]
        self.assertEqual(len(cross), 1)
        self.assertEqual(cross[0][0]["link_index"], cross[0][1]["link_index"])
        self.assertEqual({r["node_index"] for r in cross[0]}, {0, 1})
        # connections 两端对称
        for rec in instances.values():
            for link_id, conn in rec["connections"].items():
                self.assertIn(link_id, by_id)
                self.assertNotEqual(conn["instance_id"], rec["instance_id"])
        # 序号被正确推进
        self.assertEqual(self.fake.dump("/")["/next_link_index"], b"3")

    def test_dry_run_writes_nothing(self):
        before = dict(self.fake.dump("/"))
        result = self._deploy(dry_run=True)
        after = dict(self.fake.dump("/"))
        self.assertEqual(before, after)
        self.assertEqual(len(result.plan), 2 + 5 + 3)  # 2 个节点 + 5 实例 + 3 条链路记录（含跨机副本）

    def test_unknown_type_fails_before_any_write(self):
        before = dict(self.fake.dump("/"))
        with self.assertRaises(ValidationError):
            self.sn.emulation.deploy_topology([InstanceSpec("NoSuchType")], [])
        self.assertEqual(before, dict(self.fake.dump("/")))


if __name__ == "__main__":
    unittest.main()
