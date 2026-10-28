import json
import threading
import time
import unittest

from opensn import OpenSN
from opensn.schema.configs import InstanceSpec, LinkSpec
from opensn.schema.entities import Position
from opensn.tests.fake_etcd import FakeEtcd


EMU = {
    "running": False,
    "type_config": {
        "TypeA": {"image": "sat:latest", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
        "TypeB": {"image": "gs:latest", "container_envs": {}, "resource_limit": {"nano_cpu": 1, "memory_byte": 1}},
    },
}


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


class TestFlow(unittest.TestCase):
    def setUp(self):
        self.fake = FakeEtcd()
        self.fake.seed("/nodes/0", json.dumps(node_payload(0)))
        self.fake.seed("/emulation_config", json.dumps(EMU))
        self.sn = OpenSN(client=self.fake, watch=False, poll_interval=60)
        self.result = self.sn.emulation.deploy_topology(
            [InstanceSpec("TypeA", extra={"opaque_field": "0"}), InstanceSpec("TypeB", extra={"user_lat": "0"})],
            [LinkSpec(0, 1, address_infos=[{"IPV4": "10.0.0.1/30"}, {"IPV4": "10.0.0.2/30"}])],
        )
        self.sat, self.gs = self.result.instances

    def tearDown(self):
        self.sn.close()

    def test_start_sets_flags_and_parameters(self):
        link = self.sn.links.get(self.result.link_ids[0])
        before = json.loads(self.fake.dump("/node_")[f"/node_{link.node_index}/links/{link.link_id}"])["parameter"]
        link.set(delay_s=0.001, loss_ratio=0.01)
        link_record = json.loads(self.fake.dump("/node_")[f"/node_{link.node_index}/links/{link.link_id}"])
        self.assertEqual(link_record["parameter"], before)  # 还没启动：记录里的初值不变

        self.sn.emulation.start()
        link_record = json.loads(self.fake.dump("/node_")[f"/node_{link.node_index}/links/{link.link_id}"])
        self.assertTrue(link_record["enable"])
        self.assertEqual(link_record["parameter"]["delay"], 1000, "启用时应把当前参数表写回记录")
        instance_raw = json.loads(self.fake.dump("/node_")[f"/node_0/instances/{self.sat.instance_id}"])
        self.assertTrue(instance_raw["start"])
        self.assertTrue(json.loads(self.fake.dump("/")["/emulation_config"])["running"])

        self.sn.emulation.stop()
        self.assertFalse(json.loads(self.fake.dump("/")["/emulation_config"])["running"])

    def test_link_set_writes_full_table_and_both_sides(self):
        link = self.sn.links.get(self.result.link_ids[0])
        link.set(delay_s=0.002, loss_ratio=0.015, bandwidth_bps=8_000_000, connect=True)
        table = json.loads(self.fake.dump("/node_0")[f"/node_0/link_parameter/{link.link_id}"])
        self.assertEqual(table["delay"], 2000)
        self.assertEqual(table["loss"], 150)
        self.assertEqual(table["bandwidth"], 1_000_000)
        self.assertEqual(table["connect"], 1)
        link.set(connect=False)
        table = json.loads(self.fake.dump("/node_0")[f"/node_0/link_parameter/{link.link_id}"])
        self.assertEqual(table["connect"], 0)
        self.assertEqual(table["delay"], 2000, "整表提交：没改的字段不能丢")

    def test_push_config_accepts_arbitrary_content(self):
        inst = self.sn.instances.get(self.sat.instance_id)
        self.assertFalse(inst.config_exists())
        inst.push_config({"schema": "my-own", "links": {self.result.link_ids[0]: {"addr": "10.0.0.1/30"}}})
        raw = json.loads(self.fake.dump("/instance_config")[f"/instance_config/node_0/{self.sat.instance_id}"])
        self.assertEqual(raw["schema"], "my-own")
        self.assertEqual(inst.get_config()["schema"], "my-own")
        self.assertTrue(inst.config_exists())
        # 字符串内容也允许（平台不解析）
        inst.push_config("plain text")
        stored = self.fake.dump("/instance_config")[f"/instance_config/node_0/{self.sat.instance_id}"]
        self.assertEqual(stored, b"plain text")

    def test_custom_placement(self):
        fake = FakeEtcd()
        fake.seed("/nodes/0", json.dumps(node_payload(0)))
        fake.seed("/nodes/1", json.dumps(node_payload(1, free=3)))
        fake.seed("/emulation_config", json.dumps(EMU))
        sn = OpenSN(client=fake, watch=False, poll_interval=60)
        try:
            def placement(instances, nodes):
                for inst in instances:
                    inst.node_index = 1  # 用户自己的调度策略：全部塞 1 号机

            result = sn.emulation.deploy_topology(
                [InstanceSpec("TypeA"), InstanceSpec("TypeA")], [], placement=placement
            )
            self.assertEqual({i.node_index for i in result.instances}, {1})
            nodes = {k: json.loads(v) for k, v in fake.dump("/nodes").items()}
            self.assertEqual(nodes["/nodes/1"]["free_instance"], 1)
            self.assertEqual(nodes["/nodes/0"]["free_instance"], 10)
        finally:
            sn.close()

    def test_create_instance_respects_charge_flag(self):
        before = self.sn.nodes.get(0).free_instance
        inst = self.sn.instances.create("TypeA", extra={"k": "v"}, charge_capacity=False, node_index=0)
        self.assertEqual(inst.node_index, 0)
        self.assertEqual(self.sn.nodes.get(0).free_instance, before, "charge_capacity=False 不应改容量")

    def test_kv_channel(self):
        self.sn.kv.put("/my_state/plan", {"hello": "world"})
        self.assertEqual(self.sn.kv.get("/my_state/plan"), {"hello": "world"})
        self.assertEqual(self.sn.kv.get("/my_state/none", "default"), "default")

    def test_shell_accepts_custom_command(self):
        info_key = f"/node_0/webshell_info/instance_{self.sat.instance_id}"

        def daemon():
            time.sleep(0.25)
            self.fake.seed(info_key, json.dumps({"webshell_id": "x", "addr": "1.2.3.4", "port": 1, "pid": 2}))

        threading.Thread(target=daemon, daemon=True).start()
        self.sn.shell.instance(self.sat, command="sh", args=["-c", "ip addr"])
        request = json.loads(self.fake.dump("/node_0")[f"/node_0/webshell_request/instance_{self.sat.instance_id}"])
        self.assertEqual(request["command"], "sh")
        self.assertEqual(request["args"], ["-c", "ip addr"])

    def test_steps_commit_positions(self):
        steps = self.sn.steps(interval=0.05)
        step = next(steps)
        step.positions.set(self.sat, Position(latitude=0.1, longitude=0.2, altitude=550000.0))
        next(steps)  # 进入下一轮时提交上一轮
        steps.close()
        raw = json.loads(self.fake.dump("/position")[f"/position/{self.sat.instance_id}"])
        self.assertAlmostEqual(raw["latitude"], 0.1)

    def test_step_queues_link_writes(self):
        link = self.sn.links.get(self.result.link_ids[0])
        steps = self.sn.steps(interval=0.05)
        step = next(steps)
        before = self.fake.txn_count
        for i in range(3):
            self.sn.links.get(self.result.link_ids[0]).set(delay_s=0.001 + i / 1e6)
        self.assertEqual(self.fake.txn_count, before, "step 内的写应该先入队，而不是每次一条往返")
        next(steps)
        steps.close()
        table = json.loads(self.fake.dump("/node_0")[f"/node_0/link_parameter/{link.link_id}"])
        self.assertEqual(table["delay"], 1002)

    def test_shell_roundtrip(self):
        info_key = f"/node_0/webshell_info/instance_{self.sat.instance_id}"

        def daemon():
            time.sleep(0.25)
            self.fake.seed(info_key, json.dumps({"webshell_id": f"instance_{self.sat.instance_id}", "addr": "10.0.0.1", "port": 9999, "pid": 12}))

        threading.Thread(target=daemon, daemon=True).start()
        info = self.sn.shell.instance(self.sat, expire_minute=5)
        self.assertEqual(info.port, 9999)
        request = json.loads(self.fake.dump("/node_0")[f"/node_0/webshell_request/instance_{self.sat.instance_id}"])
        self.assertEqual(request["command"], "docker")
        self.assertIn("exec", request["args"])


if __name__ == "__main__":
    unittest.main()
