import ephem

class Instance:

    def __init__(self,id:str, tle:list[str]) -> None:
        self.instance_id = id
        if len(tle) < 3 :
            raise "TLE Info Not Valid"
        self.object = ephem.readtle(tle[0], tle[1], tle[2])

    def calculatePostion(self,time) -> (float,float,float):
        ephem_time = ephem.Date(time)
        self.satellite.compute(ephem_time)
        return self.satellite.sublat, self.satellite.sublong, self.satellite.elevation