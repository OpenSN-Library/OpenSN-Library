from datetime import datetime


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
    freq = 1 / period  # if the period is 1, the freq is 1Hz, if the period is 0.5, the freq is 2Hz
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
    """
    [
     [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
     [11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21], 
     [22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32], 
     [33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43], 
     [44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54], 
     [55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65]
    ]
    """

    for i in range(len(satellites)):
        y = i // orbit_satellite_num  # orbit index
        x = i % orbit_satellite_num  # satellite index

        # if number of satellite > 1, then connect to next satellite
        if orbit_satellite_num > 1:
            # connect to the next satellite in the orbit
            array = [index_2d[y][(x + 1) % orbit_satellite_num]]
        else:
            array = []
        # if not the last orbit, connect to the next orbit
        if y < orbit_num - 1:
            array.append(index_2d[y + 1][x])
        # record the inter-orbit connection
        topo[str(i)] = array
    return satellites, topo


if __name__ == "__main__":
    satellites, topo = generate_tle(8, 30, 0, 0, 0, 0.1)
    print(topo)
