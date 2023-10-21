from socket import *
from tools import get_ip_address


class DataUpdater:

    def __init__(self, host: str, port: int):
        """
        create udp socket to broadcast status
        :param host: the ip address of the interface connect to docker0
        :param port: the port to broadcast
        """
        self.addr = (host, port)
        self.udp_cli_sock = socket(AF_INET, SOCK_DGRAM)
        print(f"bind to {get_ip_address()}:35000", flush=True)
        self.udp_cli_sock.bind((get_ip_address(), 35000))
        # self.udp_cli_sock.setsockopt(SOL_SOCKET, SO_BROADCAST, 1)

    def broadcast_status(self, json_data: str):
        print(f"sending -> {json_data} to {self.addr}", flush=True)  # 写入到文件之中的需要即使的进行输出，不然在文件之中需要隔很久才能够看到
        try:
            self.udp_cli_sock.sendto(json_data.encode(), self.addr)
        except Exception as e:
            pass
        pass
