"""实体编辑能力：字段级 update、跨机链路保存、删除类接口、单端参数、kv CAS。"""

import json
import unittest

from opensn import OpenSN
from opensn import keys
from opensn.errors import NotFound, ValidationError
from opensn.schema.configs import InstanceSpec, LinkSpec
from opensn.schema.entities import Position
from opensn.tests.fake_etcd import FakeEtcd


def node_payload(index, free=10):
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
    "running": False,
    "type_config": {
        "TypeA": {"image": "a:1", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
        "TypeB": {"image": "b:1", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
    },
}


class TestEditing(unittest.TestCase):
    def setUp(self):
        self.fake = FakeEtcd()
        self.fake.seed("/nodes/0", json.dumps(node_payload(0, free=1)))
        self.fake.seed("/nodes/1", json.dumps(node_payload(1)))
        self.fake.seed("/emulation_config", json.dumps(EMU))
        self.sn = OpenSN(client=self.fake, watch=False, poll_interval=60)
        self.result = self.sn.emulation.deploy_topology(
            [InstanceSpec("TypeA"), InstanceSpec("TypeB")],
            [LinkSpec(0, 1, address_infos=[{"IPV4": "10.0.0.1/30"}, {"IPV4": "10.0.0.2/30"}])],
        )
        self.link = self.sn.links.get(self.result.link_ids[0])
        self.assertIn(len(self.sn.links.of(self.result.instances[0])), (1,))
        self.assertTrue(self.link.cross_machine, "本用例需要一条跨机链路")

    def tearDown(self):
        self.sn.close()

    def _raw(self, key):
        return self.fake.dump("/")[key.encode() if isinstance(key, str) else key] if key in self.fake.dump("/") else None

    # ---- 实例字段级编辑 ----
    def test_instance_update_merges_fields(self):
        inst = self.sn.instances.get(self.result.instance_ids[0])
        inst.update(image="new:2", extra={"k": "v"})
        raw = json.loads(self.fake.dump("/node_")[keys.instance(inst.node_index, inst.instance_id)])
        self.assertEqual(raw["image"], "new:2")
        self.assertEqual(raw["extra"], {"k": "v"})
        self.assertEqual(raw["type"], "TypeA", "未修改的字段必须保留")
        self.assertEqual(raw["node_index"], inst.node_index)
        self.assertEqual(self.sn.instances.get(inst.instance_id).image, "new:2")

    def test_instance_update_rejects_identity_and_placement(self):
        inst = self.sn.instances.get(self.result.instance_ids[0])
        with self.assertRaises(ValidationError):
            inst.update(node_index=1)
        with self.assertRaises(ValidationError):
            inst.update(instance_id="xxx")

    # ---- 链路字段级编辑（重点：跨机两份记录的 node_index 不能被覆盖） ----
    def test_link_update_preserves_per_record_node_index(self):
        old_node_indexes = {}
        for key, raw in self.fake.dump("/node_").items():
            if "/links/" in key:
                old_node_indexes[key] = json.loads(raw)["node_index"]
        self.assertEqual(set(old_node_indexes.values()), {0, 1})

        self.link.update(extra={"note": "edited"}, enable=True)
        new_node_indexes = {}
        for key, raw in self.fake.dump("/node_").items():
            if "/links/" in key:
                record = json.loads(raw)
                new_node_indexes[key] = record["node_index"]
                self.assertEqual(record["extra"], {"note": "edited"})
                self.assertTrue(record["enable"])
        self.assertEqual(new_node_indexes, old_node_indexes)

    def test_link_update_rejects_topology_fields(self):
        with self.assertRaises(ValidationError):
            self.link.update(end_infos=[])
        with self.assertRaises(ValidationError):
            self.link.update(link_index=99)

    # ---- 删除类 ----
    def test_delete_instance_config_and_position(self):
        inst = self.sn.instances.get(self.result.instance_ids[0])
        inst.push_config({"a": 1})
        cfg_key = keys.instance_config(inst.node_index, inst.instance_id)
        self.assertIsNotNone(self.sn.kv.get_raw(cfg_key))
        inst.delete_config()
        self.assertIsNone(self.sn.kv.get_raw(cfg_key))
        self.assertFalse(inst.config_exists())

        self.sn.positions.set(inst, Position(latitude=0.1))
        pos_key = keys.position(inst.instance_id)
        self.assertIsNotNone(self.sn.kv.get_raw(pos_key))
        self.sn.positions.delete(inst)
        self.assertIsNone(self.sn.kv.get_raw(pos_key))

    def test_remove_link_parameter(self):
        self.link.set(delay_s=0.001, loss_ratio=0.01)
        for node_index in (0, 1):
            self.assertIsNotNone(self.sn.kv.get_raw(keys.link_parameter(node_index, self.link.link_id)))
        self.link.remove_parameter()
        for node_index in (0, 1):
            self.assertIsNone(self.sn.kv.get_raw(keys.link_parameter(node_index, self.link.link_id)))

    def test_remove_instance_type(self):
        self.sn.emulation.remove_instance_type("TypeB")
        types = self.sn.emulation.instance_types()
        self.assertEqual(set(types), {"TypeA"})
        self.assertFalse(self.sn.emulation.status())
        with self.assertRaises(NotFound):
            self.sn.emulation.remove_instance_type("TypeB")

    # ---- 单端参数 ----
    def test_side_specific_parameters(self):
        self.link.set(delay_s=0.001, side=0)
        self.assertEqual(
            json.loads(self.sn.kv.get_raw(keys.link_parameter(0, self.link.link_id))), {"delay": 1000}
        )
        self.assertIsNone(self.sn.kv.get_raw(keys.link_parameter(1, self.link.link_id)))

        self.link.set(delay_s=0.002, side=1)
        self.assertEqual(
            json.loads(self.sn.kv.get_raw(keys.link_parameter(1, self.link.link_id))), {"delay": 2000}
        )
        self.assertEqual(
            json.loads(self.sn.kv.get_raw(keys.link_parameter(0, self.link.link_id))), {"delay": 1000},
            "side=1 不能影响另一端的参数",
        )
        self.assertEqual(self.link.get_parameters(side=0), {"delay": 1000})

        self.link.set_raw({"connect": 0, "delay": 500}, side=1)
        self.assertEqual(
            json.loads(self.sn.kv.get_raw(keys.link_parameter(1, self.link.link_id))), {"connect": 0, "delay": 500}
        )

    # ---- 自定义键的 CAS ----
    def test_kv_cas_update(self):
        self.sn.kv.put("/my_state/counter", {"n": 0})
        value, revision = self.sn.kv.get_with_revision("/my_state/counter")
        self.assertEqual(value, {"n": 0})
        self.assertIsNotNone(revision)

        self.sn.kv.cas_update("/my_state/counter", lambda current: {"n": current["n"] + 1})
        self.assertEqual(self.sn.kv.get("/my_state/counter"), {"n": 1})
        self.sn.kv.cas_update("/my_state/counter", lambda current: {"n": current["n"] + 1})
        self.assertEqual(self.sn.kv.get("/my_state/counter"), {"n": 2})
        _, new_revision = self.sn.kv.get_with_revision("/my_state/counter")
        self.assertGreater(new_revision, revision)

        # 键不存在时当前值为 None，可以直接创建
        self.sn.kv.cas_update("/my_state/new", lambda current: {"init": current is None})
        self.assertEqual(self.sn.kv.get("/my_state/new"), {"init": True})


if __name__ == "__main__":
    unittest.main()
