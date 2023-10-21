import base64
import time

import cv2
import numpy as np
from pyfiglet import Figlet
from loguru import logger
from PyInquirer import prompt
import socket

# global variable
BUFF_SIZE = 65536  # 64KB buffer size


class CompleteVideoServer:
    def __init__(self, listeningPortTmp):
        self.listeningPort = listeningPortTmp
        self.listeningDestination = "0.0.0.0"
        self.server_socket = None

    def create_server_socket_and_rcv(self):
        logger.success(f"listening at {self.listeningDestination}:{self.listeningPort}")
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, BUFF_SIZE)
        self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
        self.server_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS, bytearray([0x93, 0x8, 0x1, 0x2, 0x3, 0x4,
                                                                                       0x5, 0x6]))
        self.server_socket.bind((self.listeningDestination, self.listeningPort))
        self.server_socket_rcv()

    def server_socket_rcv(self):
        fps, st, frames_to_count, cnt = (0, 0, 20, 0)
        while True:
            packet, _ = self.server_socket.recvfrom(BUFF_SIZE)
            # "why is there a space and a slash?" - "because it's a base64 url"
            data = base64.b64decode(packet, ' /')
            npdata = np.fromstring(data, dtype=np.uint8)
            frame = cv2.imdecode(npdata, 1)
            frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
            cv2.imshow("RECEIVING VIDEO", frame)
            key = cv2.waitKey(1) & 0xFF
            if key == ord('q'):
                self.server_socket.close()
                break
            if cnt == frames_to_count:
                try:
                    fps = round(frames_to_count / (time.time() - st))
                    st = time.time()
                    cnt = 0
                except:
                    pass
            cnt += 1


if __name__ == "__main__":
    try:
        logo = "udp video server"
        logo_printer = Figlet(width=400, font="slant")
        print(logo_printer.renderText(logo))
        questions = [
            {
                "type": "input",
                "name": "listeningPort",
                "message": "Please input the listening port",
            }
        ]
        answers = prompt(questions)
        # get listening port
        listeningPort = int(answers["listeningPort"])
        # use the user input to create the complete video server
        complete_video_server = CompleteVideoServer(listeningPort)
        # create socket and rcv
        complete_video_server.create_server_socket_and_rcv()

    except KeyboardInterrupt as e:
        logger.success("udp video server exit")
