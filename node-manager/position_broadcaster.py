import json
import multiprocessing as mp
from satellite_node import worker,satellites
from const_var import *
from loguru import logger
import time
from multiprocessing import Pipe
from ground_station import ground_select,ground_stations

def generate_submission_list_for_position_broadcaster(satellite_num, cpu_count):
    if cpu_count < satellite_num:
        # each cpu handle several satellites
        submission_size = (satellite_num // cpu_count) + 1
        submission_list = []
        # [0-2] [3-5] [6-8] [9-9]
        for i in range(0, satellite_num, submission_size):
            if i + submission_size - 1 > satellite_num:
                submission_list.append((i, satellite_num - 1))
            else:
                submission_list.append((i, i + submission_size - 1))
    else:
        # each satellite is handled by one cpu
        submission_list = [(i, i) for i in range(satellite_num)]
    return submission_list


def position_broadcaster(stop_process_state, satellite_num, position_datas, updater, sending_interval):
    # create a config command and send out the command
    # ------------------------------------------------
    config_message = {"config": "set the source routing table"}
    config_str = json.dumps(config_message)
    updater.broadcast_info(config_str)
    # ------------------------------------------------

    # update position
    # ------------------------------------------------
    # 打印cpu的数量
    cpu_count = min(mp.cpu_count(), satellite_num)
    logger.info(f"cpu_count: {cpu_count}")
    # 共享数组
    res = mp.Array('f', range(3 * satellite_num), lock=False)
    first_time = True
    # 创建进程
    # 创建子任务
    submission_list = generate_submission_list_for_position_broadcaster(satellite_num, cpu_count)
    while True:
        if stop_process_state.value:
            break
        current_count = 0
        multiple_processes = []
        # 创建管道
        rcv_pipe, send_pipe = Pipe()
        for i in range(len(submission_list)):
            p = mp.Process(target=worker, args=(submission_list[i][0],
                                                submission_list[i][1],
                                                res,
                                                send_pipe))
            multiple_processes.append(p)
            p.start()
        while True:
            # rcv_int = rcv_pipe.recv()
            # current_count += rcv_int
            # # print(f"current_count: {current_count}")
            # if current_count < satellite_num:
            #     continue
            # else:
                for i in range(satellite_num):
                    node_id = 'node_' + str(i)
                    index_base = 3 * i
                    position_datas[node_id][LATITUDE_KEY] = res[index_base]
                    position_datas[node_id][LONGITUDE_KEY] = res[index_base + 1]
                    position_datas[node_id][HEIGHT_KEY] = res[index_base + 2]
                
                ground_connections = ground_select(satellites,position_datas,ground_stations)
                broadcast_data = {
                    "position_datas": position_datas,
                    "ground_connections": ground_connections
                }
                data_str = json.dumps(broadcast_data)
                updater.broadcast_info(data_str)
                
                for p in multiple_processes:
                    p.kill()
                send_pipe.close()
                rcv_pipe.close()
                break
        time.sleep(sending_interval)
    logger.success("position broadcaster process finished")


if __name__ == "__main__":
    print(generate_submission_list_for_position_broadcaster(320, 64))
