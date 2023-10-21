import socket

NETLINK_TEST = 30


def create_netlink_client():
    s = socket.socket(socket.AF_NETLINK, socket.SOCK_RAW, NETLINK_TEST)
    s.bind((0, 0))
    return s


def send_netlink_msg(s, msg):
    s.send(msg)
    data = s.recv(1024)
    return data


def main():
    s = create_netlink_client()
    msg = b"hello"
    data = send_netlink_msg(s, msg)
    print(data)
