import math

def ra2deg(ra : float) -> float:
    return ra * 180 / math.pi

def dec2ra(dec : float) -> float:
    return dec * math.pi / 180

def object2Map(obj:object):
    """对象转Dict"""
    new_obj = {}
    m = obj.__dict__
    for k in m.keys():
        v = m[k]
        if hasattr(v, "__dict__"):
            new_obj[k] = object2Map(v)
        elif isinstance(m[k],list):
            l = []
            for item in m[k]:
                if hasattr(item, "__dict__"):
                    l.append(object2Map(item))
                else:
                    l.append(item)
                new_obj[k]=l
        elif isinstance(m[k],dict):
            l = {}
            for key,item in m[k].items():
                if hasattr(item, "__dict__"):
                    l[key]=(object2Map(item))
                else:
                    l[key]=item
                new_obj[k]=l
        else:
            new_obj[k] = m[k]
    return new_obj