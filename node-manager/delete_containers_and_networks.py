import os
import time
from multiprocessing import Process, Pipe
from loguru import logger
from docker import DockerClient
from threading import Thread

def find_delete_containers():
    container_ids = []
    command_find_container = "docker ps -a | grep node-* | awk '{print $1}'"
    result_container = os.popen(command_find_container)
    for line in result_container.readlines():
        container_ids.append(("container", line.strip()))
    return container_ids


def find_delete_networks():
    network_ids = []
    command_find_network = "docker network ls | grep Network | awk '{print $1}'"
    result_network = os.popen(command_find_network)
    for line in result_network.readlines():
        network_ids.append(("network", line.strip()))
    return network_ids


def generate_submission_for_delete(containers_and_network_ids, submission_size: int):
    submission_list = []
    for i in range(len(containers_and_network_ids)):
        if i % submission_size == 0:
            submission_list.append([])
        submission_list[-1].append(containers_and_network_ids[i])
    return submission_list


def delete_containers_and_networks_submission(submission, docker_client, send_pipe):
    for single_mission in submission:
        if single_mission[0] == "container":
            docker_client.stop_satellite(single_mission[1])
            docker_client.rm_satellite(single_mission[1])
            logger.info(f"delete container {single_mission[1]}")
        elif single_mission[0] == "network":
            docker_client.rm_network(single_mission[1])
            logger.info(f"delete network {single_mission[1]}")
    send_pipe.send("finished")


def delete_containers_with_multiple_processes(docker_client: DockerClient, submission_size: int):
    start_time = time.time()
    current_finished_count = 0
    missions = find_delete_containers()
    submission_list = generate_submission_for_delete(missions, submission_size)
    rcv_pipe, send_pipe = Pipe()
    for submission in submission_list:
        process = Process(target=delete_containers_and_networks_submission, args=(submission, docker_client, send_pipe))
        process.start()
        # singleThread = Thread(target=delete_containers_and_networks_submission,
        #                       args=(submission, docker_client, send_pipe))
        # singleThread.start()
    while True:
        rcv_str = rcv_pipe.recv()
        if rcv_str == "finished":
            current_finished_count += 1
            logger.info(f"finished {current_finished_count} submission(s)")
            if current_finished_count == len(submission_list):
                rcv_pipe.close()
                send_pipe.close()
                break
    end_time = time.time()
    logger.info(f"delete containers with multiple processes cost {end_time - start_time} seconds")


def delete_networks_with_multiple_processes(docker_client: DockerClient, submission_size: int):
    start_time = time.time()
    current_finished_count = 0
    missions = find_delete_networks()
    submission_list = generate_submission_for_delete(missions, submission_size)
    rcv_pipe, send_pipe = Pipe()
    for submission in submission_list:
        process = Process(target=delete_containers_and_networks_submission, args=(submission, docker_client, send_pipe))
        process.start()
        # singleThread = Thread(target=delete_containers_and_networks_submission,
        #                       args=(submission, docker_client, send_pipe))
        # singleThread.start()
    while True:
        rcv_str = rcv_pipe.recv()
        if rcv_str == "finished":
            current_finished_count += 1
            logger.info(f"finished {current_finished_count} submission(s)")
            if current_finished_count == len(submission_list):
                rcv_pipe.close()
                send_pipe.close()
                break
    end_time = time.time()
    logger.info(f"delete networks with multiple processes cost {end_time - start_time} seconds")


if __name__ == "__main__":
    print(generate_submission_for_delete([1, 2, 3], 4))
