from configparser import ConfigParser


class Config:
    """
    Read config.ini file and override the config information.
    """

    def __init__(self, filepath: str):
        """
        :param filepath: the path of config.ini file.
        """
        self.DockerHostIP = ""
        self.DockerImageName = ""
        self.UDPPort = ""
        self.MonitorImageName = ""
        self.GroundImageName = ""
        self.GroundConfigPath = ""
        parser = ConfigParser()
        parser.read(filepath, encoding='UTF-8')
        if parser.has_option("Docker", "DockerHostIP"):
            self.DockerHostIP = parser["Docker"]["DockerHostIP"]
        if parser.has_option("Docker", "ImageName"):
            self.DockerImageName = parser["Docker"]["ImageName"]
        if parser.has_option("Docker", "UDPPort"):
            self.UDPPort = parser["Docker"]["UDPPort"]
        if parser.has_option("Docker", "MonitorImageName"):
            self.MonitorImageName = parser["Docker"]["MonitorImageName"]
        if parser.has_option("Docker", "GroundImageName"):
            self.GroundImageName = parser["Docker"]["GroundImageName"]
        if parser.has_option("Docker", "GroundConfigPath"):
            self.GroundConfigPath = parser["Docker"]["GroundConfigPath"]