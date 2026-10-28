import unittest

from opensn.errors import ValidationError
from opensn.schema.units import (
    bandwidth_bytes,
    delay_us,
    loss_raw,
    parse_size_binary,
    parse_size_decimal,
)


class TestUnits(unittest.TestCase):
    def test_parse_sizes(self):
        self.assertEqual(parse_size_decimal("50M"), 50_000_000)
        self.assertEqual(parse_size_decimal("2G"), 2_000_000_000)
        self.assertEqual(parse_size_binary("128M"), 134_217_728)
        self.assertEqual(parse_size_binary("1k"), 1024)
        self.assertEqual(parse_size_decimal(123), 123)

    def test_bad_sizes(self):
        for bad in ["0", "050", "1.5G", "50MB", "-1", "", "1e6"]:
            with self.assertRaises(ValidationError):
                parse_size_decimal(bad)
        with self.assertRaises(ValidationError):
            parse_size_binary(True)

    def test_link_units(self):
        self.assertEqual(delay_us(0.023), 23000)
        self.assertEqual(loss_raw(0.015), 150)
        self.assertEqual(bandwidth_bytes(8_000_000), 1_000_000)
        with self.assertRaises(ValidationError):
            loss_raw(1.5)
        with self.assertRaises(ValidationError):
            delay_us(-1)


if __name__ == "__main__":
    unittest.main()
