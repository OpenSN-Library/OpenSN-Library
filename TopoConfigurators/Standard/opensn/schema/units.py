"""单位换算与取值校验。

daemon 内部的单位（见 task-spec/ref/06-etcd-only-protocols.md、附录 B）：
- 链路参数 delay：微秒整数，daemon 下发 netem 时还会 +1µs
- 链路参数 loss：万分之一（150 = 1.5%）
- 链路参数 bandwidth：字节/秒（tbf 的 rate）
- 位置：弧度 + 米
- 资源限制：请求里的 "50M" 字符串，nano_cpu 用十进制、memory_byte 用 1024 进制
"""

import re

from ..errors import ValidationError

_NUMBER_RE = re.compile(r"^([1-9][0-9]*)([KMGTkmgt]?)$")

_DEC = {"k": 10 ** 3, "m": 10 ** 6, "g": 10 ** 9, "t": 10 ** 12}
_BIN = {"k": 2 ** 10, "m": 2 ** 20, "g": 2 ** 30, "t": 2 ** 40}


def _parse_number(value, scale, kind: str) -> int:
    if isinstance(value, bool):
        raise ValidationError(kind, value, "bool is not a size")
    if isinstance(value, int):
        if value < 0:
            raise ValidationError(kind, value, "must be >= 0")
        return value
    if not isinstance(value, str):
        raise ValidationError(kind, value, 'expected int or "50M" style string')
    m = _NUMBER_RE.match(value)
    if not m:
        raise ValidationError(
            kind, value, 'expected ^[1-9][0-9]*([KMGTkmgt])?$ (0, decimals and "050" are invalid)'
        )
    base = int(m.group(1))
    suffix = m.group(2).lower()
    if suffix:
        base *= scale[suffix]
    return base


def parse_size_decimal(value) -> int:
    """nano_cpu：K/M/G/T 按 10 进制（daemon ParseDecNumber）。"""
    return _parse_number(value, _DEC, "nano_cpu")


def parse_size_binary(value) -> int:
    """memory_byte：K/M/G/T 按 1024 进制（daemon ParseBinNumber）。"""
    return _parse_number(value, _BIN, "memory_byte")


def delay_us(seconds) -> int:
    """秒 -> 微秒整数。"""
    try:
        us = int(round(float(seconds) * 1_000_000))
    except (TypeError, ValueError):
        raise ValidationError("delay_s", seconds, "expected seconds as float or timedelta")
    if us < 0:
        raise ValidationError("delay_s", seconds, "must be >= 0")
    return us


def loss_raw(ratio) -> int:
    """丢包比例（0.015 = 1.5%）-> 万分之一整数（150）。"""
    try:
        raw = int(round(float(ratio) * 10000))
    except (TypeError, ValueError):
        raise ValidationError("loss_ratio", ratio, "expected ratio in [0, 1]")
    if raw < 0 or raw > 10000:
        raise ValidationError("loss_ratio", ratio, "must be in [0, 1]")
    return raw


def bandwidth_bytes(bps) -> int:
    """bit/s -> 字节/秒。"""
    try:
        v = int(round(float(bps) / 8.0))
    except (TypeError, ValueError):
        raise ValidationError("bandwidth_bps", bps, "expected bits per second")
    if v < 0:
        raise ValidationError("bandwidth_bps", bps, "must be >= 0")
    return v


def rad_from_deg(deg: float) -> float:
    return float(deg) / 180.0 * 3.141592653589793


def deg_from_rad(rad: float) -> float:
    return float(rad) * 180.0 / 3.141592653589793
