import multiprocessing
import os
from loguru import logger


def start_multicast_process():
    os.system("python example_multicast.py")


def start_multi_unicast_process():
    os.system("python example_multi_unicast.py")


def pre_calculate(destinationNodeNumTmp, destinationNodeList):
    for i in range(0, destinationNodeNumTmp):
        destination_node_id = destinationNodeList[i]
        # we need to pre insert some routes
        # --------- call caller.py to insert route --------
        logger.info("start to insert route")
        os.chdir("../satellite_node")
        command = f"python caller.py {os.getenv('NODE_ID')[5:]} {destination_node_id} > /dev/null"
        os.system(command)
        # return to the original path
        os.chdir("../udp_live_stream")
        logger.info("finish to insert route")


def pre_calculate_multicast(destinationNodeNum, destinationNodeList):
    logger.info("start to insert route")
    os.chdir("../satellite_node")
    command = f"python caller.py {os.getenv('NODE_ID')[5:]} "
    for i in range(destinationNodeNum):
        command += f"{destinationNodeList[i]} "
    # command += " > /dev/null"
    print(command)
    os.system(command)
    os.chdir("../udp_live_stream")
    logger.info("finish inserting route")


if __name__ == "__main__":
    unicast_destination_node_num = 2
    unicast_destination_node_list = [47, 46]
    multicast_destination_node_num = 2
    multicast_destination_node_list = [29, 30]
    pre_calculate(unicast_destination_node_num, unicast_destination_node_list)
    pre_calculate_multicast(multicast_destination_node_num, multicast_destination_node_list)
    multiprocessing.Process(target=start_multicast_process).start()
    multiprocessing.Process(target=start_multi_unicast_process).start()
