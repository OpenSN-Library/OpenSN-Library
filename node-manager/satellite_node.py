import ephem
from datetime import datetime
from global_var import satellites
from loguru import logger


# 用来进行位置计算的线程
def worker(range_start: int, range_end: int, res, send_pipe):
    """
    Calculate the position of the satellite
    :param send_pipe:
    :param res:
    :param range_start: the start index of the satellites that this thread should calculate
    :param range_end: the end index of the satellites that this thread should calculate
    """
    # calculated satellite nums
    calculated_satellites_num = range_end - range_start + 1
    # calculate the position of the satellites
    now = datetime.utcnow()
    for i in range(range_start, range_end + 1):
        index_base = 3 * i
        # logger.info("%d %d %d %d"%(len(res),index_base,len(satellites),i))
        res[index_base], res[index_base + 1], res[index_base + 2] = satellites[i].get_next_position(now)
    send_pipe.send(calculated_satellites_num)


class SatelliteNode:

    def __init__(self, tle_info: tuple, node_id: str, container_id: str):
        self.orbit = tle_info[0][5:].split('_')[0]
        self.position = tle_info[0][5:].split('_')[1]
        self.satellite = ephem.readtle(tle_info[0], tle_info[1], tle_info[2])
        self.node_id = node_id
        self.container_id = container_id
        self.topo = []
        self.host_ip = ''
        self.subnet_ip = {}  # {subnet_str: interface}

    def __str__(self):
        return self.node_id

    def get_next_position(self, time_now):
        """
        Get the next position of the satellite
        :param time_now: the time now
        :return: the next position of the satellite
        """
        ephem_time = ephem.Date(time_now)
        self.satellite.compute(ephem_time)
        return self.satellite.sublat, self.satellite.sublong, self.satellite.elevation


if __name__ == "__main__":
    pass
