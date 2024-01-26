

IPV4_BASE = [10,0,0,0]

def array_add(base: list,delta: int) -> list:
    ret = [base[i] for i in range(len(base))]
    for i in range(len(base)-1,-1,-1):
        if delta == 0 :
            break
        delta = delta + base[i]
        ret[i] = delta % 256
        delta = delta // 256
    return ret

def alloc_ipv4(prefix_len: int) -> list:
    global IPV4_BASE
    delta = 1<<(32-prefix_len)
    subnet = []
    for i in range(delta):
        subnet.append(array_add(IPV4_BASE,i))
    IPV4_BASE = array_add(IPV4_BASE,delta)
    print(subnet)
    return subnet

def format_ipv4(ip:list, prefix_len: int) -> str:
    if len(ip) < 4:
        raise "invalid ip"
    return "%d.%d.%d.%d/%d"%(ip[0],ip[1],ip[2],ip[3],prefix_len)
    