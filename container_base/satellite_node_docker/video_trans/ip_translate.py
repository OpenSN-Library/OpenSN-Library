
def ip2str(ip: int) -> str:
    return '.'.join([str(ip >> (i << 3) & 0xFF) for i in range(4)[::-1]])

def str2ip(ip: str) -> int:
    return sum([int(ip.split('.')[i]) << ((3 - i) << 3) for i in range(4)])