import socket
from loguru import logger
from PyInquirer import prompt
import time
from pyfiglet import Figlet


class CompleteServer:

    def __init__(self):
        self.bind_port = None
        self.udp_server_socket = None
        self.udp_server_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

    def create_udp_server_socket(self):
        self.udp_server_socket.bind(("0.0.0.0", self.bind_port))
        while True:
            data, address = self.udp_server_socket.recvfrom(1024)
            # convert bytes to string
            data = data.decode()
            # if receive q then quit
            if data == "":
                raise KeyboardInterrupt
            elif data.startswith("time:"):
                print(f"received from client:", address, "time elapsed:", time.time() - float(data[5:]))
            else:
                print(f"received from client {address}: {data}")


if __name__ == "__main__":
    try:
        logo = "udp server"
        logo_printer = Figlet(width=600, font="slant")
        print(logo_printer.renderText(logo))
        questions = [
            {
                "type": "input",
                "name": "bindPort",
                "message": "Please input the port:",
            }
        ]
        answers = prompt(questions)
        logger.success("udp server started")
        completeServer = CompleteServer()
        completeServer.bind_port = int(answers["bindPort"])
        completeServer.create_udp_server_socket()

    except KeyboardInterrupt as e:
        logger.success("udp server stopped")
