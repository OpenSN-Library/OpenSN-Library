import time
from socket import *
from loguru import logger


class DataUpdater:

    def __init__(self, host: str, eth_ip: str, port: int):
        self.addr = (host, port)
        self.eth_ip = eth_ip
        self.udp_cli_sock = socket(AF_INET, SOCK_DGRAM)
        self.udp_cli_sock.bind((eth_ip, 36000))
        self.udp_cli_sock.setsockopt(SOL_SOCKET, SO_BROADCAST, 1)

    def broadcast_info(self, json_data: str):
        try:
            # why come to errno 22? invalid argument because of the first reason
            # broadcast address is not in the same subnet
            # another reason is that the port is not in the range of 0-65535
            # another reason is that the port is already in use
            # how to judge the port is in use? use netstat -anp | grep 30000
            # if the port is in use
            logger.info("sending broadcast info")
            self.udp_cli_sock.sendto(json_data.encode(), self.addr)
        except Exception as e:
            logger.error(f"{self.addr, self.eth_ip}")


if __name__ == "__main__":
    updater = DataUpdater("<broadcast>", "172.17.0.1", 30000)
    with open("error.txt", "r") as f:
        for i in range(30):
            json_str = f.read()
            updater.broadcast_info(json_str)
            time.sleep(3)
