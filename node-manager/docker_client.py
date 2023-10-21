import logging
import docker
from const_var import *
from subnet_allocator import SubnetAllocator, ip2str

class DockerClient:

    def __init__(self, image_name: str, host_ip: str, ground_image_name: str):
        self.client = docker.from_env()
        self.image_name = image_name
        self.ground_image_name = ground_image_name
        self.allocator = SubnetAllocator(29)
        self.host_ip = host_ip

    
    def create_satellite(self, node_id: str, port: str, satellite_num: int, successful_init) -> str:
        # add --cap-add=NET_ADMIN
        index = int(node_id.split('_')[1])
        if successful_init:
            container_info = self.client.containers.run(self.image_name, detach=True, environment=[
                'NODE_ID=' + node_id,
                'HOST_IP=' + self.host_ip,
                'BROAD_PORT=' + port,
                'THRESHOLD=' + str(5.0),
                "SATELLITE_NUM=" + str(satellite_num),
                "MONITOR_IP=" + "172.17.0.2",
                "DISPLAY=unix:0.0",
                "GDK_SCALE",
                "GDK_DPI_SCALE",
            ], cap_add=['NET_ADMIN'], name=node_id, volumes=[
                VOLUME1, VOLUME2,V_EDIT], privileged=True, ports={'8765/tcp': 30000+index})
        else:
            container_info = self.client.containers.run(self.image_name, detach=True, environment=[
                'NODE_ID=' + node_id,
                'HOST_IP=' + self.host_ip,
                'BROAD_PORT=' + port,
                'THRESHOLD=' + str(5.0),
                "SATELLITE_NUM=" + str(satellite_num),
                "MONITOR_IP=" + "172.17.0.1",
                "DISPLAY=unix:0.0",
                "GDK_SCALE",
                "GDK_DPI_SCALE",
            ], cap_add=['NET_ADMIN'], name=node_id, volumes=[
                VOLUME1, VOLUME2,V_EDIT], privileged=True)
        return container_info.id

    def create_ground_container(self, node_id_g: str):
        container_info = self.client.containers.run(self.ground_image_name, detach=True, environment=[
                'NODE_ID=' + node_id_g,
                'HOST_IP=' + self.host_ip,
                "MONITOR_IP=" + "172.17.0.2",
                "DISPLAY=unix:0.0",
            ], cap_add=['NET_ADMIN'], name=node_id_g, volumes=[VOLUME2,V_EDIT], privileged=True,command=["tail","-f","/dev/null"])
        return container_info.id
    

    def stop_satellite(self, container_id: str):
        self.client.containers.get(container_id).stop()

    def rm_satellite(self, container_id: str):
        self.client.containers.get(container_id).remove(force=True)

    def rm_network(self, network_id: str):
        self.client.networks.get(network_id).remove()

    def connect_node(self, container_id: str, network_id: str, alias_name: str):
        container_object = self.client.containers.get(container_id)
        self.client.networks.get(network_id).connect(container_object, aliases=[alias_name])

    def disconnect_node(self, container_id: str, network_id: str):
        container_object = self.client.containers.get(container_id)
        self.client.networks.get(network_id).disconnect(container_object)

    def pull_image(self):
        image = self.client.pull(self.image_name)
        return image.id

    def create_network_with_ipam_config(self, network_index, ipam_config):
        net = self.client.networks.create("Network_" + str(network_index), driver='bridge', ipam=ipam_config)
        return net.id

    def create_network(self, network_id: str):
        net = None
        for i in range(100):
            try:
                subnet_ip = self.allocator.alloc_local_subnet()
                subnet_ip_str = ip2str(subnet_ip)
                gateway_str = ip2str(subnet_ip + 1)
                ipam_pool = docker.types.IPAMPool(subnet='%s/29' % subnet_ip_str, gateway=gateway_str)
                ipam_config = docker.types.IPAMConfig(pool_configs=[ipam_pool])
                net = self.client.networks.create(network_id, driver='bridge', ipam=ipam_config)
                break
            except Exception as e:
                print(e)
                logging.info("create network error, retried %d time" % i)
        if net is None:
            raise "create net work error"
        return net.id,subnet_ip

    def get_container_interfaces(self, container_id: str):
        ans = []
        free_bit = []
        container_info = self.client.containers.get(container_id)
        nets = container_info.attrs["NetworkSettings"]["Networks"]
        for net_name in nets.keys():
            ans.append(nets[net_name]["IPAddress"])
            free_bit.append(int(nets[net_name]["IPPrefixLen"]))
        return ans, free_bit


    def exec_cmd(self,container_id: str, cmd: list):
        self.client.containers.get(container_id).exec_run(tty=False,cmd=cmd)

if __name__ == '__main__':
    cli = DockerClient('aaa', 'bbb')
    resp = cli.get_container_interfaces("6ad80cb5be8ba205027c157814b0f47eda11c2a635862fba367316df8a2720d0")
    print(resp)
