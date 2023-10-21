import socket
from loguru import logger

addr_replace = {}

class UDPSender:

    def __create_sender_socket(self,port: int) -> socket.socket :
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.bind(('', port))
        s.settimeout(2)
        self.retry_time = 16
        return s

    def __init__(self,port) -> None:
        self.port : int = port
        self.socket : socket.socket = self.__create_sender_socket(port)

    def send_data(self, data: bytes, t_addr:list, t_port: int) -> None:
        logger.info("Send %d Bytes by UDP to %s"%(len(data), t_addr[0]))
        i = 0
        while i < self.retry_time:
            real_addr = t_addr[i%len(t_addr)]
            if real_addr in addr_replace.keys():
                real_addr = addr_replace[real_addr]
            try:
                self.socket.sendto(data, (real_addr, t_port))
                i += 1
                self.socket.recv(1024)
                if real_addr != t_addr[0]:
                    addr_replace[t_addr[0]] = real_addr
                break
            except Exception as e:
                print("ACK Timeout, Retrying %s"%real_addr)
            

class TCPSender:
    def __create_sender_socket(self,port: int) -> socket.socket :
        s = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
        s.bind(('', port))
        s.settimeout(2)
        self.retry_time = 16
        return s
    
    def __init__(self,port:int, t_addr: str, t_port: int) -> None:
        self.port : int = port
        self.socket : socket.socket = self.__create_sender_socket(port)
        self.t_addr = t_addr
        self.t_port = t_port
        self.socket.connect((self.t_addr,self.t_port))

    def send_data(self,data: bytes) -> None:
        logger.info("Send %d Bytes by TCP..."%(len(data)))
        i = 0
        while i < self.retry_time:
            try:
                i += 1
                self.socket.send(data)
                break
            except Exception as e:
                print("ACK Timeout, Retrying")