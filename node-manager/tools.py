
def ip_to_subnet(ip: str, prefix_len: int) -> str:
    # print(ip, prefix_len)
    sections = ip.split('.', -1)
    if len(sections) < 4:
        raise Exception("Incorrect Format")
    ip_int = (int(sections[0]) << 24) + (int(sections[1]) << 16) + (int(sections[2]) << 8) + int(sections[3])

    shift_num = 32 - prefix_len
    ip_int = ip_int >> shift_num
    ip_int = ip_int << shift_num
    return "%d.%d.%d.%d" % (
        ip_int // (1 << 24),
        ip_int // (1 << 16) % 256,
        ip_int // (1 << 8) % 256,
        ip_int % 256
    )


if __name__ == "__main__":
    print(ip_to_subnet("172.17.0.3", 16))
    print(ip_to_subnet("172.17.8.3", 24))
    print(ip_to_subnet("172.17.9.3", 8))
    print(ip_to_subnet("172.17.0.3", 15))
