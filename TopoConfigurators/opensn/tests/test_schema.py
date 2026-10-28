import base64
import socket
import unittest

from opensn.schema.entities import Instance, Link, Node, Position


class TestSchema(unittest.TestCase):
    def test_instance_roundtrip(self):
        raw = {
            "instance_id": "3f2a9c1d",
            "name": "MyNodeType_3f2a9c1d",
            "type": "MyNodeType",
            "image": "img",
            "extra": {"user_key": "x"},
            "device_need": {"fplink": {"dev_name": "fplink", "need_num": 1, "is_mutex": True}},
            "resource": {"nano_cpu": 50000000, "memory_byte": 134217728},
            "node_index": 0,
            "connections": {"abc": {"link_id": "abc", "instance_id": "peer", "instance_type": "MyNodeType", "end_node_index": 1}},
            "start": True,
        }
        inst = Instance.from_json(raw)
        out = inst.to_dict()
        self.assertEqual(out["device_need"]["fplink"]["need_num"], 1)
        self.assertNotIn("device_info", out)
        self.assertEqual(out["connections"]["abc"]["end_node_index"], 1)
        self.assertEqual(out["resource"]["nano_cpu"], 50000000)

    def test_link_arrays_are_fixed_length_2(self):
        link = Link.from_json({"link_id": "abc", "end_infos": [{"instance_id": "a"}], "address_infos": []})
        self.assertEqual(len(link.end_infos), 2)
        self.assertEqual(len(link.address_infos), 2)
        out = link.to_dict()
        self.assertIsInstance(out["end_infos"], list)
        self.assertEqual(len(out["end_infos"]), 2)
        self.assertEqual(len(out["address_infos"]), 2)

    def test_node_base64_addresses(self):
        v4 = base64.b64encode(socket.inet_aton("10.0.0.5")).decode()
        mac = base64.b64encode(bytes.fromhex("aabbccddeeff")).decode()
        node = Node.from_json({"node_index": 1, "l_3_addr_v_4": v4, "l_2_addr": mac, "cpu_core": 8})
        self.assertEqual(node.ipv4, "10.0.0.5")
        self.assertEqual(node.mac, "aa:bb:cc:dd:ee:ff")
        self.assertEqual(node.cpu_core, 8)

    def test_position_degrees(self):
        pos = Position.from_degrees(90.0, 180.0, 550000.0)
        lat, lon, alt = pos.degrees()
        self.assertAlmostEqual(lat, 90.0)
        self.assertAlmostEqual(lon, 180.0)
        self.assertEqual(alt, 550000.0)


if __name__ == "__main__":
    unittest.main()
