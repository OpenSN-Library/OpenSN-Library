import socket
from const_var import TCP_SERVER_PORT


class TcpClient:
    def __init__(self, ip, port):
        self.ip = ip
        self.port = port
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        print(f"connect to {self.ip}:{self.port}")
        # add flag to the socket
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
        self.socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                               bytearray([0x93, 0x6, 0x1, 0x2, 0x3, 0x4]))
        self.socket.connect((self.ip, self.port))

    def send(self, send_data):
        self.socket.send(send_data)

    def receive(self):
        return self.socket.recv(1024)

    def close(self):
        self.socket.close()


if __name__ == "__main__":
    tcp_client = None
    try:
        print("start tcp client.py")
        server_ip = input("please input the server ip about to connect:")
        tcp_client = TcpClient(server_ip, TCP_SERVER_PORT)
        while True:
            data = input("input data (input q or quit to exit):")
            if data == "q" or data == "quit":
                raise KeyboardInterrupt
            tcp_client.send(data.encode())
            print(f"client.py received from server echo: {tcp_client.receive().decode()}")
    except KeyboardInterrupt:
        print("tcp client.py close")
        if tcp_client is not None:
            tcp_client.close()
