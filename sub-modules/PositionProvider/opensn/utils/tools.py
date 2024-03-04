import math

def ra2deg(ra : float) -> float:
    return ra * 180 / math.pi

def dec2ra(dec : float) -> float:
    return dec * math.pi / 180



def object2dict(obj: any) -> any:
    if hasattr(obj,"__dict__"):
        ret = {}
        for k,v in obj.__dict__.items():
            ret[k] = object2dict(v)
        return ret
    elif isinstance(obj,list):
        ret = []
        for item in obj:
            ret.append(object2dict(item))
        return ret
    elif isinstance(obj,dict):
        ret = {}
        for k,v in obj.items():
            ret[k] = object2dict(v)
        return ret
    else:
        return obj