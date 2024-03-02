from datetime import datetime
import json,itertools

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


def generate_tle(orbit_num: int, orbit_satellite_num: int, latitude, longitude, delta, period) -> (list, dict):
    satellites = []
    index_2d = []
    topo = {}
    freq = 1 / period
    line_1 = "1 00000U 23666A   %02d%012.8f  .00000000  00000-0 00000000 0 0000"
    line_2 = "2 00000  90.0000 %08.4f 0000011   0.0000 %8.4f %11.8f00000"
    year2, day = get_year_day(datetime.now())

    for i in range(orbit_num):
        start_latitude = latitude + delta * i
        start_longitude = longitude + 180 * i / orbit_num
        index_1d = []
        for j in range(orbit_satellite_num):
            this_latitude = start_latitude + 360 * j / orbit_satellite_num
            this_line_1 = line_1 % (year2, day)
            this_line_2 = line_2 % (start_longitude, this_latitude, freq)
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
        if y < orbit_num - 1:
            array.append(index_2d[y + 1][x])
        topo[str(i)] = array
    return satellites, topo


if __name__ == "__main__":
    sat, topos = generate_tle(21, 21, 0, 0, 5, 0.05)
    tle_fd = open("tle/tle.tle","w")
    flat_sat =[]
    for tle in sat:
        flat_sat.append(tle[0]+'\n')
        flat_sat.append(tle[1]+'\n')
        flat_sat.append(tle[2]+'\n')
    tle_fd.writelines(flat_sat)
    tle_fd.close()
    topo_fd = open("tle/topo.json","w")
    topo_fd.write(json.dumps(topos))
    topo_fd.close()
