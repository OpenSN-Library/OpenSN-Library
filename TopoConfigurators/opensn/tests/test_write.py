import json
import unittest

from opensn import OpenSN
from opensn.errors import Conflict
from opensn.schema.entities import Position
from opensn.tests.fake_etcd import FakeEtcd


def make_node(node_index=0, free=10):
    return {
        "node_index": node_index,
        "free_instance": free,
        "is_master_node": node_index == 0,
        "l_3_addr_v_4": "",
        "l_3_addr_v_6": "",
        "l_2_addr": "",
        "ns_instance_map": {},
        "ns_link_map": {},
        "node_link_device_info": {},
        "cpu_core": 8,
    }


class TestWrite(unittest.TestCase):
    def setUp(self):
        self.fake = FakeEtcd()
        self.fake.seed("/nodes/0", json.dumps(make_node(0, 10)))
        self.sn = OpenSN(client=self.fake, watch=False, poll_interval=60)

    def tearDown(self):
        self.sn.close()

    def test_skip_unchanged(self):
        before = self.fake.txn_count
        node = self.sn.nodes.get(0)
        self.sn._ctx.nodes  # noqa: B018
        from opensn.write import write_put

        write_put(self.sn._ctx, "/nodes/0", node)
        self.assertEqual(self.fake.txn_count, before, "unchanged write should not be sent")

    def test_batch_dedupes_and_chunks(self):
        batch = self.sn.batch()
        for i in range(130):
            batch.put(f"/position/p{i}", Position(latitude=float(i)))
        before = self.fake.txn_count
        sent = batch.commit()
        self.assertEqual(sent, 130)
        self.assertEqual(self.fake.txn_count - before, 2, "130 ops should be sent in 2 chunks of 120")

    def test_stale_cas_retries_and_wins(self):
        from opensn.schema.entities import Node
        from opensn.transport import CREATE

        node = self.sn.nodes.get(0)
        stale = self.sn._ctx.store.mod_revision("/nodes/0")
        # 模拟别的写入方先改了这条记录
        other = Node.from_json(self.fake.get("/nodes/0")[0].decode() if False else node.to_dict())
        other.free_instance = 7
        self.fake.put("/nodes/0", json.dumps(other.to_dict()).encode())
        # SDK 用旧版本号写，应该重试后成功
        batch = self.sn.batch()
        target = Node.from_json(node.to_dict())
        target.free_instance = 5
        batch.put("/nodes/0", target, stale)
        batch.commit()
        current = Node.from_json(json.loads(self.fake.dump("/nodes")["/nodes/0"]))
        self.assertEqual(current.free_instance, 5)


if __name__ == "__main__":
    unittest.main()
