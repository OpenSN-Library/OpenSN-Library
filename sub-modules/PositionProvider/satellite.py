import ephem
import json
import datetime
from link import LinkBase
from instance import Instance
from tools import ra2deg
from const_var import TYPE_SATELLITE
from threading import RLock

'''
type Position struct {
    Latitude  float64
    Longitude float64
    Altiutde  float64
}
'''




class Satellite(Instance):

    def __init__(self,id:str, tle:list[str],orbit_index:int,satellite_index:int,ns:str,node_index:int) -> None:
        Instance.__init__(self,id,TYPE_SATELLITE,ns,node_index)
        if len(tle) < 3 :
            raise "TLE Info Not Valid"
        self.__object = ephem.readtle(tle[0], tle[1], tle[2])
        self.orbit_index = orbit_index
        self.satellite_index = satellite_index
        

    def calculate_postion(self,time:datetime.datetime) -> (float,float,float):
        ephem_time = ephem.Date(time)
        self.__object.compute(ephem_time)
        self.latitude = self.__object.sublat
        self.longitude = self.__object.sublong
        self.altitude = self.__object.elevation

    
