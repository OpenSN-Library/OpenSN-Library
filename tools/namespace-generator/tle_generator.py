from datetime import datetime
from type_model_def import EmulationTypeConfig, TopologyLink, TopologyInstance, TopologyConfig
from instance_types import EX_TLE0_KEY,EX_TLE1_KEY,EX_TLE2_KEY,EX_ORBIT_INDEX,EX_SATELLITE_INDEX,EX_AREA_KEY,EX_AREA_X,EX_AREA_Y,EX_TOTAL_AREA_X,EX_TOTAL_AREA_Y
from address_type import LINK_V4_ADDR_KEY
import json
def get_year_day(now_time: datetime) -> (int, float):
    year = now_time.year
    day = float(now_time.microsecond)
    day /= 1000
    day += now_time.second
    day /= 60
    day += now_time.minute
    day /= 60
    day += now_time.hour
    day /= 24
    day += (now_time - datetime(year, 1, 1)).days

    return year % 100, day

def assign_area(satellite_grid: list[list[TopologyInstance]], area_x: int, area_y: int):
    area_x_toal = len(satellite_grid) // area_x
    area_y_total = len(satellite_grid[0]) // area_y
    for i in range(len(satellite_grid)):
        for j in range(len(satellite_grid[i])):
            grid_x = i // area_x
            grid_y = j // area_y
            satellite_grid[i][j].extra[EX_AREA_KEY] = f'0.0.0.{grid_x*area_y_total+grid_y+1}'
            satellite_grid[i][j].extra[EX_AREA_X] = str(grid_x)
            satellite_grid[i][j].extra[EX_AREA_Y] = str(grid_y)
            satellite_grid[i][j].extra[EX_TOTAL_AREA_X] = str(area_x_toal)
            satellite_grid[i][j].extra[EX_TOTAL_AREA_Y] = str(area_y_total)


def assign_address(satellite_grid: list[list[TopologyInstance]], link_array: list[TopologyLink]):
    area_address_map = {}
    for link in link_array:
        node_x_0 = link.end_indexes[0] // len(satellite_grid[0])
        node_y_0 = link.end_indexes[0] % len(satellite_grid[0])
        area_index_0 = satellite_grid[node_x_0][node_y_0].extra[EX_AREA_KEY].split('.')[-1]
        node_x_1 = link.end_indexes[1] // len(satellite_grid[0])
        node_y_1 = link.end_indexes[1] % len(satellite_grid[0])
        area_index_1 = satellite_grid[node_x_1][node_y_1].extra[EX_AREA_KEY].split('.')[-1]
        if area_index_0 == area_index_1:
            if area_index_0 not in area_address_map:
                area_address_map[area_index_0] = 0
            addr_index = area_address_map[area_index_0]
            link.address_infos[0][LINK_V4_ADDR_KEY] = f"10.{area_index_0}.{addr_index//256}.{addr_index%256 + 1}/30"
            link.address_infos[1][LINK_V4_ADDR_KEY] = f"10.{area_index_0}.{addr_index//256}.{addr_index%256 + 2}/30"
            area_address_map[area_index_0] += 4
        else:
            if area_index_0 not in area_address_map:
                area_address_map[area_index_0] = 0
            if area_index_1 not in area_address_map:
                area_address_map[area_index_1] = 0
            addr_index_0 = area_address_map[area_index_0]
            addr_index_1 = area_address_map[area_index_1]
            link.address_infos[0][LINK_V4_ADDR_KEY] = f"10.{area_index_0}.{addr_index_0//256}.{addr_index_0%256 + 1}/30"
            link.address_infos[1][LINK_V4_ADDR_KEY] = f"10.{area_index_1}.{addr_index_1//256}.{addr_index_1%256 + 1}/30"
            area_address_map[area_index_0] += 4
            area_address_map[area_index_1] += 4


def str_checksum(line: str) -> int:
    sum_num = 0
    for c in line:
        if c.isdigit():
            sum_num += int(c)
        elif c == '-':
            sum_num += 1
    return sum_num % 10


def area2line(y: int, x: int, x_limit: int, y_limit: int) -> int:
    y_true = (y + y_limit) % y_limit
    x_true = (x + x_limit) % x_limit
    return y_true * x_limit + x_true


def generate_tle(orbit_num: int, orbit_satellite_num: int, all_start_latitude, all_start_longitude, orbit_angle, delta_percent, period) -> (list, dict):
    satellites = []
    index_2d = []
    topo = {}
    freq = 1 / period
    line_1 = "1 00000U 23666A   %02d%012.8f  .00000000  00000-0 00000000 0 0000"
    line_2 = "2 00000  %02.4f %08.4f 0000011   0.0000 %8.4f %11.8f00000"
    year2, day = get_year_day(datetime.now())
    total_longitude = 180
    if abs(orbit_angle) < 80:
        total_longitude = 360
    delta = 360 / orbit_satellite_num * delta_percent
    for i in range(orbit_num):
        start_latitude = all_start_latitude + delta * i
        start_longitude = all_start_longitude + total_longitude * i / orbit_num
        index_1d = []
        for j in range(orbit_satellite_num):
            this_latitude = start_latitude + 360 * j / orbit_satellite_num
            this_line_1 = line_1 % (year2, day)
            this_line_2 = line_2 % (orbit_angle, start_longitude, this_latitude, freq)
            index_1d.append(len(satellites))
            satellites.append(
                [
                    "NODE_%d_%d" % (i, j),
                    this_line_1 + str(str_checksum(this_line_1)),
                    this_line_2 + str(str_checksum(this_line_2))
                ]
            )
        index_2d.append(index_1d)

    for i in range(len(satellites)):
        y = i // orbit_satellite_num
        x = i % orbit_satellite_num
        if orbit_satellite_num > 1:
            array = [index_2d[y][(x + 1) % orbit_satellite_num]]
        else:
            array = []
        if abs(orbit_angle) < 80 or y < orbit_num - 1:
            array.append(index_2d[(y + 1) % orbit_num][x])
        topo[str(i)] = array
    return satellites, topo

area_size_list = ["1x9","1x18","2x9","2x18","4x9","4x18","8x9","8x18"]

if __name__ == "__main__":
    grid_x = 40
    grid_y = 18
    for area_size in area_size_list:
        area_x = int(area_size.split('x')[0])
        area_y = int(area_size.split('x')[1])
        start_longitude = 0
        start_latitude = 0
        orbit_angle = 90
        delta_percent = 1 / grid_x
        period = 1 / 13.1507
        emu_config :dict[str,EmulationTypeConfig] = {
            "Satellite": EmulationTypeConfig("docker.io/realssd/satellite-router-area", {}, "50M", "128M").__dict__
        }
        sat, topos = generate_tle(grid_x, grid_y, start_longitude, start_latitude, orbit_angle, delta_percent, period)
        node_grid = []
        for i in range(grid_x):
            array = []
            for j in range(grid_y):
                array.append(TopologyInstance("Satellite", {
                    EX_TLE0_KEY: sat[area2line(i, j, grid_y, grid_x)][0],
                    EX_TLE1_KEY: sat[area2line(i, j, grid_y, grid_x)][1],
                    EX_TLE2_KEY: sat[area2line(i, j, grid_y, grid_x)][2],
                    EX_ORBIT_INDEX: str(i),
                    EX_SATELLITE_INDEX: str(j),
                }))
            node_grid.append(array)
        links = []
        for l0_str,l1_list in topos.items():
            for l1 in l1_list:
                links.append(TopologyLink("vlink", [int(l0_str),l1],{},[{},{}],{}))
        assign_area(node_grid, area_x, area_y)
        assign_address(node_grid, links)
        emu_config_file = open("emu_config.json", "w")
        emu_config_file.write(json.dumps(emu_config))
        emu_config_file.close()
        topology_config = TopologyConfig()
        for i in range(grid_x):
            for j in range(grid_y):
                topology_config.instances.append(node_grid[i][j])
        
        for link in links:
            topology_config.links.append(link)

        topology_config_file = open("topology_config_oneweb_%d_%d.json"%(area_x,area_y), "w")
        print(f"Generate topology_config_oneweb_{area_x}_{area_y}.json")
        topology_config_file.write(topology_config.toJson())