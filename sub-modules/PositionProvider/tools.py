import math

def ra2deg(ra : float) -> float:
    return ra * 180 / math.pi

def dec2ra(dec : float) -> float:
    return dec * math.pi / 180