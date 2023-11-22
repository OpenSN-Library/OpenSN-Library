import ephem
import json

'''
type Position struct {
    Latitude  float64
    Longitude float64
    Altiutde  float64
}

'''


class Instance:

    def __init__(self,id:str, tle:list[str]) -> None:
        self.__instance_id = id
        if len(tle) < 3 :
            raise "TLE Info Not Valid"
        self.__object = ephem.readtle(tle[0], tle[1], tle[2])
        self.latitude = 0.0
        self.longitude = 0.0
        self.altitude = 0.0
        

    def calculatePostion(self,time) -> (float,float,float):
        ephem_time = ephem.Date(time)
        self.__object.compute(ephem_time)
        self.latitude = self.satellite.sublat
        self.longitude = self.satellite.sublong
        self.altitude = self.satellite.elevation
    
class InstanceEncoder(json.JSONEncoder):
    def default(self, obj : Instance):
        return {
            "latitude" : obj.latitude,
            "longitude" : obj.longitude,
            "altitude" : obj.altitude
        }

if __name__ == "__main__":
    sat = {
        "a": Instance('114514',[
            "BEIDOU 3 ",
            "1 36287U 10001A   21187.60806788 -.00000272  00000-0  00000-0 0  9992",
            "2 36287   1.9038  47.2796 0005620  82.9429 153.9116  1.00269947 42045"
        ]),
        "b": Instance('114514',[
            "BEIDOU 3 ",
            "1 36287U 10001A   21187.60806788 -.00000272  00000-0  00000-0 0  9992",
            "2 36287   1.9038  47.2796 0005620  82.9429 153.9116  1.00269947 42045"
        ])
    }

    print(json.dumps(sat,cls=InstanceEncoder))