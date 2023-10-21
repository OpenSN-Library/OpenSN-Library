import time
import requests
import json
from loguru import logger

port = 8080


def init_monitor(image_name: str, client, udp_port: int) -> bool:
    result = True
    try:
        client.client.containers.run(image_name, detach=True, environment=[
            'UDP_PORT=' + str(udp_port),
        ], ports={'%d/tcp' % port: port}, name="satellite-monitor")
    except Exception as e:
        logger.error("monitor already exists")
        result = False
    finally:
        return result


def connect_monitor():
    while True:
        try:
            requests.post(url='http://127.0.0.1:%d/api/satellite/print' % port)
            break
        except Exception as e:
            logger.warning("connect monitor error, retrying...")
            time.sleep(1)
    # print connect success
    logger.success("connect monitor success")


def set_monitor(raw_payload: list, ground_payload: dict, stop_process_state, interval: int):
    while True:
        try:
            if stop_process_state.value:
                break
            payload = {
                'total': len(raw_payload),
                'items': raw_payload
            }
            # 进行post请求的发送
            requests.post(url='http://127.0.0.1:%d/api/satellite/list' % port, data=json.dumps(payload))
            requests.post(url='http://127.0.0.1:%d/api/ground/position' % port, data=json.dumps(ground_payload))
            time.sleep(interval)
        except Exception as e:
            logger.warning("connect monitor error, retrying...")
            time.sleep(1)
    logger.success("set monitor exit")
