import base64
import time
from ctypes import c_bool
import cv2
import imutils
import os
import socket
from enum import Enum
from PyInquirer import prompt
from loguru import logger
from pyfiglet import Figlet
import pexpect as px
import multiprocessing

# fixed parameters

BUFF_SIZE = 65536  # 64KB buffer size
WIDTH = 400
UNICAST_FRAME_TITLE = "UNICAST TRANSMITTING VIDEO"
MULTICAST_FRAME_TITLE = "MULTICAST TRANSMITTING VIDEO"


# TransmissionPattern
class TransmissionPattern(Enum):
    unicast = 1
    multicast = 2
    multi_unicast = 3


class CompleteVideoClient:
    lipsinMulticastAlreadyStartClient = multiprocessing.Value(c_bool, False)
    multipleUnicastAlreadyStartClient = multiprocessing.Value(c_bool, False)

    def __init__(self, enableLipsinTmp,
                 transmissionPatternTmp,
                 destinationPortTmp,
                 videoPathTmp):
        self.unicast_client_video_socket = None  # unicast client video socket
        self.multicast_client_video_socket = None  # multicast client video socket
        self.enableLipsin = enableLipsinTmp  # if enable lipsin
        self.transmissionPattern = transmissionPatternTmp  # unicast or multicast
        self.destinationPort = destinationPortTmp  # destination port
        self.video_path = videoPathTmp  # video path
        self.destinationNodeNum = None  # destination node number
        self.destinationNodeId = None  # destination node id
        self.destinationIp = None  # destination ip
        self.vid = None  # video id
        self.destinationNodeList = []  # destination node list
        self.destinationIPList = []  # destination ip list
        self.unicastSocketList = []  # unicast socket list

    def create_socket_and_send_video(self):
        if self.transmissionPattern == TransmissionPattern.unicast:
            # get destination id
            destination_node_id = int(input("please input the destination node id: "))
            # --------- whether with or without lipsin we need to call caller.py to insert route --------
            logger.info("start to insert route")
            os.chdir("../satellite_node")
            command = f"python caller.py {os.getenv('NODE_ID')[5:]} {destination_node_id} > /dev/null"
            os.system(command)
            # return to the original path
            os.chdir("../udp_live_stream")
            logger.info("finish to insert route")
            # --------- call caller.py to insert route --------
            # find dest ip by user input id
            self.destinationIp = self.find_dest_ip_by_id(destination_node_id)
            # create unicast video socket
            self.unicast_client_video_socket = self.create_unicast_client_video_socket()
            # send video by unicast
            self.send_video_with_unicast_socket(self.unicast_client_video_socket, self.destinationIp,
                                                self.destinationPort)
        elif self.transmissionPattern == TransmissionPattern.multicast:
            # create multicast video socket
            self.multicast_client_video_socket = self.create_multicast_client_video_socket()
            # send video by multicast
            self.send_video_with_multicast_packet(self.multicast_client_video_socket, self.destinationPort)
        elif self.transmissionPattern == TransmissionPattern.multi_unicast:
            # create multi unicast video socket
            self.create_multi_unicast_video_socket()
            # we need to create a thread for each unicast socket
            self.send_video_with_multi_unicast_socket()

    def create_multi_unicast_video_socket(self):
        self.destinationNodeNum = int(input("Please input destination node number: "))
        for i in range(0, self.destinationNodeNum):
            destination_node_id = int(input("please input the destination node id: "))
            # we need to pre insert some routes
            # --------- call caller.py to insert route --------
            logger.info("start to insert route")
            os.chdir("../satellite_node")
            command = f"python caller.py {os.getenv('NODE_ID')[5:]} {destination_node_id} > /dev/null"
            os.system(command)
            # return to the original path
            os.chdir("../udp_live_stream")
            logger.info("finish to insert route")
            # --------- call caller.py to insert route --------
            self.destinationNodeList.append(destination_node_id)
            # get destination ip by id
            ip_tmp = self.find_dest_ip_by_id(destination_node_id)
            # add to the IP list
            self.destinationIPList.append(ip_tmp)
            # create unicast socket for ip
            unicast_socket_tmp = self.create_unicast_client_video_socket()
            # add the unicast socket tmp to the unicast socket list
            self.unicastSocketList.append(unicast_socket_tmp)

    def create_unicast_client_video_socket(self):
        unicast_client_video_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        unicast_client_video_socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, BUFF_SIZE)
        if self.enableLipsin:
            unicast_client_video_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
            unicast_client_video_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                                                   bytearray([0x94, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5,
                                                              0x6]))
        return unicast_client_video_socket

    def create_multicast_client_video_socket(self):
        final_byte_array = bytearray([0x94, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x93, 0x8])
        for i in range(0, self.destinationNodeNum):
            destination_node_id = self.destinationNodeList[i]
            final_byte_array.append(destination_node_id)
        # now we gonna insert routes
        # --------- whether with or without lipsin we need to call caller.py to insert route --------

        # --------- call caller.py to insert route --------
        if self.destinationNodeNum < 6:
            for i in range(0, 6 - self.destinationNodeNum):
                final_byte_array.append(0xff)
        multicast_client_video_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        # self.multicast_client_video_socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, BUFF_SIZE)
        if self.enableLipsin:
            multicast_client_video_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
            multicast_client_video_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS, final_byte_array)
        return multicast_client_video_socket

    def find_dest_ip_by_id(self, dest_node_id):
        """
        find destination ip address by destination satellite node id
        :param dest_node_id: destination satellite node id
        :return: find destination ip address
        """
        path_of_exe = "/netlink_test_userspace/build/netlink_test_userspace"
        start_node_id = os.getenv("NODE_ID")[5:]
        process = px.spawn(path_of_exe, encoding="utf-8")
        process.sendline(f"lipsin find dest ip {start_node_id}|{dest_node_id}")
        process.expect("from kernel:.*\n")
        result = process.match.group(0)[:-1]
        result = result[(result.find(":") + 1):].strip()
        if result == "node id corresponding not found":
            logger.error("node id corresponding not found")
            return
        process.sendline("q")
        process.expect(px.EOF)
        return result

    def find_dest_ip_by_id_with_question(self):
        """
        find destination ip address by self destination satellite node id
        :return:
        """
        path_of_exe = "/netlink_test_userspace/build/netlink_test_userspace"
        questions_for_find_dest = [
            {
                "type": "input",
                "name": "destNodeId",
                "message": "Please input the destination node id: ",
            }
        ]
        answers_for_unicast = prompt(questions_for_find_dest)
        start_node_id = os.getenv("NODE_ID")[5:]
        dest_node_id = answers_for_unicast["destNodeId"]
        process = px.spawn(path_of_exe, encoding="utf-8")
        process.sendline(f"lipsin find dest ip {start_node_id}|{dest_node_id}")
        process.expect("from kernel:.*\n")
        result = process.match.group(0)[:-1]
        result = result[(result.find(":") + 1):].strip()
        if result == "node id corresponding not found":
            logger.error("node id corresponding not found")
            return
        self.destinationIp = result
        self.destinationNodeId = dest_node_id
        process.sendline("q")
        process.expect(px.EOF)

    def send_video_with_unicast_socket_in_multi_process(self, video_path_tmp, socket_tmp, destination_ip_tmp,
                                                        destination_port_tmp):
        cnt, st, frames_to_count, fps = (0, 0, 20, 0)
        display_send_interface = False
        if not CompleteVideoClient.multipleUnicastAlreadyStartClient.value:
            CompleteVideoClient.multipleUnicastAlreadyStartClient.value = True
            display_send_interface = True
        vid = cv2.VideoCapture(video_path_tmp)  # read video
        while vid.isOpened():
            _, frame = vid.read()
            frame = imutils.resize(frame, width=WIDTH)
            encoded, buffer = cv2.imencode('.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
            message = base64.b64encode(buffer)
            socket_tmp.sendto(message, (destination_ip_tmp, destination_port_tmp))
            # modified
            if display_send_interface:
                frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
                cv2.imshow(UNICAST_FRAME_TITLE, frame)
            key = cv2.waitKey(1) & 0xFF
            if key == ord('q'):
                socket_tmp.close()
                break
            if cnt == frames_to_count:
                try:
                    fps = round(frames_to_count / (time.time() - st))
                    st = time.time()
                    cnt = 0
                except Exception:
                    pass
            cnt += 1

    def send_video_with_unicast_socket(self, socket_tmp, destination_ip_tmp, destination_port_tmp):
        fps, st, frames_to_count, cnt = (0, 0, 20, 0)
        self.vid = cv2.VideoCapture(self.video_path)  # read video
        while self.vid.isOpened():
            _, frame = self.vid.read()
            frame = imutils.resize(frame, width=WIDTH)
            encoded, buffer = cv2.imencode('.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
            message = base64.b64encode(buffer)
            socket_tmp.sendto(message, (destination_ip_tmp, destination_port_tmp))
            frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
            cv2.imshow(UNICAST_FRAME_TITLE, frame)
            key = cv2.waitKey(1) & 0xFF
            if key == ord('a'):
                socket_tmp.close()
                break
            if cnt == frames_to_count:
                try:
                    fps = round(frames_to_count / (time.time() - st))
                    st = time.time()
                    cnt = 0
                except Exception:
                    pass
            cnt += 1

    def send_video_with_multicast_packet(self, socket_tmp, destination_port_tmp):
        # cnt is the number of frames
        # st is the start time equals to 0
        # frames_to_count is the number of frames to count the fps
        # fps is the frame per second
        fps, st, frames_to_count, cnt = (0, 0, 20, 0)
        display_send_interface = False
        if not CompleteVideoClient.multipleUnicastAlreadyStartClient.value:
            CompleteVideoClient.multipleUnicastAlreadyStartClient.value = True
            display_send_interface = True
        self.vid = cv2.VideoCapture(self.video_path)  # read video
        while self.vid.isOpened():
            _, frame = self.vid.read()
            frame = imutils.resize(frame, width=WIDTH)
            encoded, buffer = cv2.imencode('.jpg', frame, [cv2.IMWRITE_JPEG_QUALITY, 80])
            message = base64.b64encode(buffer)
            socket_tmp.sendto(message, ("192.168.0.1", destination_port_tmp))
            if display_send_interface:
                frame = cv2.putText(frame, 'FPS: ' + str(fps), (10, 40), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 0, 255), 2)
                cv2.imshow(MULTICAST_FRAME_TITLE, frame)
            key = cv2.waitKey(1) & 0xFF
            if key == ord('q'):
                socket_tmp.close()
                break
            # how many count of frames to calculate the fps
            if cnt == frames_to_count:
                try:
                    fps = round(frames_to_count / (time.time() - st))
                    st = time.time()
                    cnt = 0
                except Exception:
                    pass
            cnt += 1

    def send_video_with_multi_unicast_socket(self):
        for i in range(len(self.unicastSocketList)):
            # when process the multiple client , we need to start only one client interface
            multiprocessing.Process(target=self.send_video_with_unicast_socket_in_multi_process,
                                    args=(self.video_path,
                                          self.unicastSocketList[i],
                                          self.destinationIPList[i],
                                          self.destinationPort)).start()

    @staticmethod
    def change_transmission_pattern_string_to_enum(transmission_pattern_string):
        if transmission_pattern_string == "unicast":
            return TransmissionPattern.unicast
        elif transmission_pattern_string == "multicast":
            return TransmissionPattern.multicast
        else:
            return TransmissionPattern.multi_unicast


def start_multicast_routing():
    enableLipsin = True
    transmissionPattern = TransmissionPattern.multicast
    destinationPort = 38789
    video_path = "movie_clip.mp4"
    complete_video_client = CompleteVideoClient(enableLipsin,
                                                transmissionPattern,
                                                destinationPort,
                                                video_path)
    complete_video_client.destinationNodeNum = 2
    complete_video_client.destinationNodeList = [29, 30]
    complete_video_client.create_socket_and_send_video()


if __name__ == "__main__":
    try:
        logo = "udp video client"
        logo_printer = Figlet(width=400, font="slant")
        print(logo_printer.renderText(logo))
        start_multicast_routing()
    except Exception:
        pass

# if __name__ == "__main__":
#     try:
#         logo = "udp video client"
#         logo_printer = Figlet(width=400, font="slant")
#         print(logo_printer.renderText(logo))
#         first_level_questions = [
#             {
#                 "type": "list",
#                 "name": "enableLipsin",
#                 "message": "Enable Lipsin? (y/n) if n only support unicast",
#                 "choices": ["y", "n"],
#             }
#         ]
#         answers_for_first_level = prompt(first_level_questions)
#         # get enableLipsin
#         enableLipsin = True if answers_for_first_level["enableLipsin"] == "y" else False
#         transmissionPatternChoices = None
#         if enableLipsin:
#             transmissionPatternChoices = ["unicast", "multicast"]
#         else:
#             transmissionPatternChoices = ["unicast", "multi_unicast"]
#         second_level_questions = [
#             {
#                 "type": "list",
#                 "name": "transmissionPattern",
#                 "message": "select transmission pattern",
#                 "choices": transmissionPatternChoices,
#             },
#             {
#                 "type": "input",
#                 "name": "destinationPort",
#                 "message": "Please input the destination port",
#             },
#             {
#                 "type": "input",
#                 "name": "video_path",
#                 "message": "Please input the video_path (default:movie_clip.mp4)",
#             }
#         ]
#         answers_for_second_level = prompt(second_level_questions)
#         transmissionPatternChoice = CompleteVideoClient. \
#             change_transmission_pattern_string_to_enum(answers_for_second_level["transmissionPattern"])
#         # get destination port
#         destinationPort = int(answers_for_second_level["destinationPort"])
#         # get the video_path
#         video_path = answers_for_second_level["video_path"]
#         if video_path == "":
#             video_path = "movie_clip.mp4"
#         # create complete video client
#         complete_video_client = CompleteVideoClient(enableLipsin,
#                                                     transmissionPatternChoice,
#                                                     destinationPort,
#                                                     video_path)
#         # create socket and send video
#         complete_video_client.create_socket_and_send_video()
#
#     except KeyboardInterrupt as e:
#         logger.success("udp video client exit")
