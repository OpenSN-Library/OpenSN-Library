import os
import socket
import time
from enum import Enum
from PyInquirer import prompt
from loguru import logger
import pexpect as px
from pyfiglet import Figlet


class TransmissionPattern(Enum):
    unicast = 1
    multicast = 2


class CompleteClient:

    def __init__(self, enableLipsinTmp, transmissionPatternTmp, destinationPortTmp):
        """
        create CompleteClient object
        :param enableLipsinTmp: if enable lipsin
        :param transmissionPatternTmp:  unicast or multicast
        :param destinationPortTmp: destination port
        """
        self.enableLipsin = enableLipsinTmp
        self.transmissionPattern = transmissionPatternTmp
        self.destinationPort = destinationPortTmp
        self.destinationIp = None
        self.lipsin_unicast_socket = None
        self.lipsin_multicast_socket = None
        self.destination_node_number = None

    def createSocket(self):
        """
        create socket (the socket could be unicast or multicast socket)
        :return:
        """
        if self.transmissionPattern == TransmissionPattern.unicast:
            # find dest ip by id
            self.find_dest_ip_by_id()
            # create socket
            self.create_lipsin_unicast_socket()
            # use the socket to send data
            self.send_packet_with_unicast_socket()
        elif self.transmissionPattern == TransmissionPattern.multicast:
            # set a fixed destination address
            self.destinationIp = "192.168.0.1"
            # create socket
            self.create_lipsin_multicast_socket()
            # use the socket to send data
            self.send_packet_with_multicast_socket()
        else:
            logger.error("transmission pattern is not supported")

    def create_lipsin_unicast_socket(self):
        """
        create lipsin unicast socket
        :return:
        """
        self.lipsin_unicast_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        if self.enableLipsin:
            self.lipsin_unicast_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
            # the set sock opt length is set to 8 bytes
            self.lipsin_unicast_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                                                  bytearray([0x94, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5,
                                                             0x6]))

    def create_lipsin_multicast_socket(self):
        """
        create lipsin multicast packet
        :return:
        """
        final_byte_array = bytearray([0x94, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x93, 0x8])
        self.lipsin_multicast_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.lipsin_multicast_socket.setsockopt(socket.SOL_SOCKET, socket.SO_DEBUG, 1)
        self.destination_node_number = int(input("please input the destination node number: "))
        if self.destination_node_number > 6:
            logger.error("the destination node number is too large")
            return
        else:
            for i in range(0, self.destination_node_number):
                # put the destination into the final byte array
                destination_node_id = int(input("please input the destination node id:"))
                final_byte_array.append(destination_node_id)
            # here we are going to call caller.python
            # --------- call caller.py to insert route --------
            logger.info("start to insert route")
            os.chdir("../satellite_node")
            command = f"python caller.py {os.getenv('NODE_ID')[5:]} "
            for i in range(0, self.destination_node_number):
                command += f"{final_byte_array[10 + i]} "
            # command += " > /dev/null"
            os.system(command)
            logger.info("finish inserting route")
            # --------- call caller.py to insert route --------
            if self.destination_node_number < 6:
                # fill the rest of the final byte array with 0
                for i in range(0, 6 - self.destination_node_number):
                    final_byte_array.append(0xff)
            self.lipsin_multicast_socket.setsockopt(socket.IPPROTO_IP, socket.IP_OPTIONS,
                                                    final_byte_array)

    def send_packet_with_unicast_socket(self):
        """
        send packet with lipsin unicast socket
        :return:
        """
        while True:
            data = input("input message: ")
            if data == "q" or data == "quit":
                raise KeyboardInterrupt
            elif data == "time":
                time_str = "time:" + str(time.time())
                self.lipsin_unicast_socket.sendto(time_str.encode(), (self.destinationIp, self.destinationPort))
            else:
                self.lipsin_unicast_socket.sendto(data.encode(), (self.destinationIp, self.destinationPort))

    def send_packet_with_multicast_socket(self):
        """
        send packet with lipsin multicast socket
        :return:
        """
        while True:
            data = input("input message: ")
            if data == "q" or data == "quit":
                raise KeyboardInterrupt
            elif data == "time":
                time_str = "time:" + str(time.time())
                self.lipsin_multicast_socket.sendto(time_str.encode(), (self.destinationIp, self.destinationPort))
            else:
                self.lipsin_multicast_socket.sendto(data.encode(), (self.destinationIp, self.destinationPort))

    def find_dest_ip_by_id(self):
        """
        find destination ip address by destination satellite node id
        :return:
        """
        path_of_exe = "/netlink_test_userspace/build/netlink_test_userspace"
        questions_for_unicast = [
            {
                "type": "input",
                "name": "destNodeId",
                "message": "Please input the destination node id: ",
            }
        ]
        answers_for_unicast = prompt(questions_for_unicast)
        start_node_id = os.getenv("NODE_ID")[5:]
        dest_node_id = answers_for_unicast["destNodeId"]
        # --------- call caller.py to insert route --------
        logger.info("start to insert route")
        os.chdir("../satellite_node")
        command = f"python caller.py {os.getenv('NODE_ID')[5:]} {dest_node_id}"
        os.system(command)
        logger.info("finish to insert route")
        # --------- call caller.py to insert route --------
        process = px.spawn(path_of_exe, encoding="utf-8")
        process.sendline(f"lipsin find dest ip {start_node_id}|{dest_node_id}")
        process.expect("from kernel:.*\n")
        result = process.match.group(0)[:-1]
        result = result[(result.find(":") + 1):].strip()
        if result == "node id corresponding not found":
            logger.error("node id corresponding not found")
            return
        self.destinationIp = result
        process.sendline("q")
        process.expect(px.EOF)


def pre_store_calculated_source_routing_items():
    already_calculated_file_path = f"/configuration/state/already_calculated_{os.getenv('NODE_ID', None)}.txt"
    is_file_exist = os.path.exists(already_calculated_file_path)
    if is_file_exist:
        with open(already_calculated_file_path, "r") as f:
            is_already_calculated = f.read().strip()
            if is_already_calculated == "true":
                is_already_calculated = True
    else:
        is_already_calculated = False
        with open(already_calculated_file_path, "w") as f:
            f.write("true")
    if not is_already_calculated:
        # we need to call netlink to transmit the source routing items to linux kernel
        logger.info("pre store calculated source routing items start")
        os.chdir("../satellite_node")
        os.system("python caller.py > /dev/null")
        logger.info("pre store calculated source routing items finished")


if __name__ == "__main__":
    # pre_store_calculated_source_routing_items()
    try:
        logo = "udp client"
        logo_printer = Figlet(width=600, font="slant")
        print(logo_printer.renderText(logo))
        questions_first_level = [
            # question1 : whether to enable lipsin
            {
                "type": "list",
                "name": "enableLipsin",
                "message": "Enable Lipsin? (y/n) if n only support unicast",
                "choices": ["y", "n"]
            },

        ]
        answers_first_level = prompt(questions_first_level)
        # get enableLipsin choice
        enableLipsinChoice = True if answers_first_level["enableLipsin"] == "y" else False
        transmissionPatternChoices = None
        if enableLipsinChoice:
            transmissionPatternChoices = ["unicast", "multicast"]
            options = "unicast/multicast"
        else:
            transmissionPatternChoices = ["unicast"]
            options = "unicast"
        questions_second_level = [
            # question2 : transmission pattern
            {
                "type": "list",
                "name": "transmissionPattern",
                "message": f"Select transmission pattern ({options})",
                "choices": transmissionPatternChoices
            },
            # question3 : destination port
            {
                "type": "input",
                "name": "destPort",
                "message": "Please input the destination port:",
            }
        ]
        answers_second_level = prompt(questions_second_level)
        # not enableLipsin only unicast is supported
        transmissionPatternChoice = None
        transmissionPatternChoice = TransmissionPattern.unicast if answers_second_level[
                                                                       "transmissionPattern"] == "unicast" \
            else TransmissionPattern.multicast
        # get destination port
        destinationPort = int(answers_second_level["destPort"])
        # use the user input parameters to create socket and send message
        completeClient = CompleteClient(enableLipsinChoice, transmissionPatternChoice, destinationPort)
        # use the client to create socket
        completeClient.createSocket()
    except KeyboardInterrupt:
        logger.success("udp client exit")
