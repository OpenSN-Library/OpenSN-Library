import json
import time
import unittest

from opensn import OpenSN
from opensn.errors import Conflict
from opensn.schema.entities import Position
from opensn.tests.fake_etcd import FakeEtcd


class _AlwaysFailEtcd(FakeEtcd):
    def transaction(self, compare=None, success=None, failure=None):
        self.txn_count += 1
        from opensn.tests.fake_etcd import _Response

        return _Response(False, self._revision)


class TestStore(unittest.TestCase):
    def test_watch_updates_local_mirror(self):
        fake = FakeEtcd()
        fake.seed("/nodes/0", json.dumps({"node_index": 0, "free_instance": 5, "cpu_core": 8}))
        sn = OpenSN(client=fake, watch=True)
        try:
            self.assertEqual(sn.nodes.get(0).free_instance, 5)
            fake.seed("/nodes/0", json.dumps({"node_index": 0, "free_instance": 3, "cpu_core": 8}))
            deadline = time.time() + 2
            while time.time() < deadline:
                if sn.nodes.get(0).free_instance == 3:
                    break
                time.sleep(0.05)
            self.assertEqual(sn.nodes.get(0).free_instance, 3)
        finally:
            sn.close()

    def test_permanent_conflict_raises(self):
        fake = _AlwaysFailEtcd()
        fake.seed("/position/x", json.dumps({"latitude": 0, "longitude": 0, "altitude": 0}))
        sn = OpenSN(client=fake, watch=False, poll_interval=60, cas_retries=2)
        try:
            with self.assertRaises(Conflict):
                with sn.batch() as batch:
                    batch.put("/position/x", Position(latitude=1.0), 1)
        finally:
            sn.close()


if __name__ == "__main__":
    unittest.main()
