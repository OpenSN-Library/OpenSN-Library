"""SDK 异常。全部继承 :class:`OpenSNError`。

设计原则：
- 读不到实体 -> NotFound
- 写之前校验失败 -> ValidationError（不会带病写入）
- 版本冲突重试耗尽 -> Conflict
- etcd 连不上/超时/无节点 -> NotReady
- etcd 返回错误 -> EtcdError
- 平台不支持（例如没有 etcd 客户端） -> Unsupported
"""


class OpenSNError(Exception):
    """所有 SDK 异常的基类。"""


class NotFound(OpenSNError):
    def __init__(self, entity: str, key: str = ""):
        self.entity = entity
        self.key = key
        super().__init__(f"{entity} not found" + (f": {key}" if key else ""))


class ValidationError(OpenSNError):
    def __init__(self, field: str, value=None, reason: str = ""):
        self.field = field
        self.value = value
        self.reason = reason
        msg = f"invalid {field}"
        if value is not None:
            msg += f"={value!r}"
        if reason:
            msg += f": {reason}"
        super().__init__(msg)


class Conflict(OpenSNError):
    def __init__(self, key: str, tries: int):
        self.key = key
        self.tries = tries
        super().__init__(f"version conflict on {key} after {tries} tries")


class NotReady(OpenSNError):
    pass


class EtcdError(OpenSNError):
    def __init__(self, message: str, code: int = 0):
        self.code = code
        self.message = message
        super().__init__(f"etcd error code={code}: {message}")


class Unsupported(OpenSNError):
    pass
