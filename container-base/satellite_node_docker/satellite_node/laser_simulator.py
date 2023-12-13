from random import expovariate,seed
import time, math
import logging

latitude_threshold = 66.5
interface_down_time = {}
interface_state_map = {}
initialized = False
INTER_ORBIT = "inter-orbit-link"  
LINK_DOWN_DURATION = 5

def init_interface_conn(interface_map):
    global initialized
    if initialized :
        return
    seed(time.time())
    
    for link in interface_map.keys():
        interface_down_time[link] = 0
        interface_state_map[link] = True
    print(interface_state_map)
    print(interface_down_time)
    initialized = True

def judge_connect(latitude, interface_map, link_failure_rate) -> dict:
    print("latitude is %f, threshold is %f" % (latitude, latitude_threshold), flush=True)
    error_res = {}

    interface_failure_rate = 1 - math.sqrt(1 - link_failure_rate)
    poisson_lambda = interface_failure_rate / (LINK_DOWN_DURATION * (1 - interface_failure_rate))

    now = int(time.time())

    for if_name in interface_state_map.keys():
        if now > interface_down_time[if_name] + LINK_DOWN_DURATION:
            error_res[if_name] = True
            interface_down_time[if_name] = now + int(round(expovariate(poisson_lambda)))
            print("Set interface %s type %s connection state up cause error"%(if_name,interface_map[if_name]))
        elif now > interface_down_time[if_name]:
            error_res[if_name] = False
            print("Set interface %s type %s connection state down cause error"%(if_name,interface_map[if_name]))
        else:
            error_res[if_name] = True
            print("Set interface %s type %s connection state up cause error"%(if_name,interface_map[if_name]))

    if abs(latitude) > latitude_threshold:
        for if_name in interface_state_map.keys():
            if interface_map[if_name] == INTER_ORBIT:
                print("Set interface %s type %s connection state down"%(if_name,interface_map[if_name]))
                interface_state_map[if_name] = False
            else:
                interface_state_map[if_name] = True and error_res[if_name]
    else:
        for if_name in interface_state_map.keys():
            if interface_map[if_name] == INTER_ORBIT:
                print("Set interface %s type %s connection state up"%(if_name,interface_map[if_name]))
                interface_state_map[if_name] = True and error_res[if_name]
            else:
                interface_state_map[if_name] = True and error_res[if_name]

    return interface_state_map
