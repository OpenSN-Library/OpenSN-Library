from random import normalvariate

latitude_threshold = 66.5


def judge_connect(latitude, threshold):
    print("latitude is %f, threshold is %f" % (latitude, latitude_threshold), flush=True)
    # zhf modified : in source routing we will let it always be true
    if abs(latitude) > latitude_threshold:
        return False
    else:
        # this line
        if abs(normalvariate(0, 1)) > threshold:
            return False
    return True
