import socket
from const_var import TCP_SERVER_IP, TCP_SERVER_PORT


class TcpServer:
    def __init__(self, ip, port):
        self.ip = ip
        self.port = port
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
        self.socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                               bytearray([0x93, 0x6, 0x1, 0x2, 0x3, 0x4]))
        self.socket.bind((self.ip, self.port))
        self.socket.listen(1)
        self.client, self.address = self.socket.accept()
        self.client.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
        self.client.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                               bytearray([0x93, 0x6, 0x1, 0x2, 0x3, 0x4]))
        print(f"connect from {self.address[0]}:{self.address[1]}")

    def receive(self):
        return self.client.recv(1024)

    def send(self, data_send):
        self.client.send(data_send)

    def close(self):
        self.client.close()
        self.socket.close()


if __name__ == "__main__":
    print("start tcp server")
    tcp_server = None
    try:
        tcp_server = TcpServer(TCP_SERVER_IP, TCP_SERVER_PORT)
        while True:
            data = tcp_server.receive()
            # convert bytes to string
            data = data.decode()
            if data == "":
                print("tcp client.py close")
                raise KeyboardInterrupt
            print("received from client.py: %s" % data)
            tcp_server.send(data.encode())
    except KeyboardInterrupt as e:
        print("tcp server close")
        if tcp_server is not None:
            tcp_server.close()
