# OpenSN

> To Be Done

[中文版本](readme_zh.md)

[TOC]

## 1. Statement

## 2. Environment Preparation

OpenSN relies on Docker Engine for container management. Ensure Docker Engine is installed before deploying OpenSN.

The Docker Engine version required is `Docker version 24.0.5`.

For Ubuntu systems, use the following command to install Docker via apt:
```bash
sudo apt install docker.io
```
Alternatively, visit [Install Docker Engine](https://docs.docker.com/engine/install/) for more installation options.

## 3. Installation

Before simulation, prepare the following materials:

1. Daemon executable files and configuration files
2. Dependency service container images
3. Topology rule controllers
4. Images needed for virtual network nodes

### 3.1 Using Precompiled Installation Packages

OpenSN provides precompiled packages for the `Linux-amd64` architecture, which include:

1. Daemon executable files and configuration files
2. Dependency service container images
3. Example topology rule controllers
4. Example images for virtual network nodes
5. OpenSN Python control SDK
6. Automated constellation configuration generation scripts

### 3.2 Compiling from Source

The directory structure of the OpenSN repository is as follows:
```bash
OpenSN
├── Makefile
├── TopoConfigurators # Example topology rule controllers
├── container-base # Directory for building example container network images
├── daemon # Subdirectory for daemon code
├── dependencies # Directory for building dependency service images
├── doc # Development documentation [TBD]
├── readme.md
├── readme_zh.md
├── tools # Automation tools
└── ui # Subdirectory for web interface code
```

## 4. Starting OpenSN

### 4.1 Kernel Configuration Before Starting

To ensure the virtual network operates correctly, configure the following kernel parameters before starting OpenSN:
```bash
# Increase the limit on inotify instances, as real-time monitoring in containers requires creating inotify instances
echo 4096 > /proc/sys/fs/inotify/max_user_instances
# Increase ARP table recycling thresholds to prevent ARP table convergence issues in large-scale virtual networks
echo 8192 > /proc/sys/net/ipv4/neigh/default/gc_thresh1 
echo 16384 > /proc/sys/net/ipv4/neigh/default/gc_thresh2 
echo 32768 > /proc/sys/net/ipv4/neigh/default/gc_thresh3

# Adjust the above values based on the number of containers started on a single machine
# Reconfigure these parameters after every system reboot
```

### 4.2 Starting Dependency Services

OpenSN will work in conjunction with three dependency services: Etcd, InfluxDB, and CodeServer. The roles of these services are as follows:

* Etcd: Facilitates information exchange between machine nodes and submodules
* InfluxDB: Stores monitoring data for the virtual environment's load
* CodeServer: Web interface editor for the Topology Configurator, optional

### 4.3 Starting the Daemon

After completing steps 4.1 and 4.2, run the `NodeDaemon` executable file with administrator privileges to start OpenSN on the machine.

The `NodeDaemon` file relies on the `config/config.json` file in the same directory for configuration.

Modify the `config.json` file according to the deployment environment before starting the `NodeDaemon` file.

```json
{
  "app": { // Configuration related to NodeDaemon startup
    "is_servant": false, // Whether to start as a slave node, set to false to start as a master node
    "master_address": "172.16.208.128", // IP address of the master node, effective only for slave nodes for interaction with the master node
    "listen_port": 8080, // HTTP service listening port, effective only for the master node
    "enable_monitor": true, // Whether to enable load monitoring service for the simulation environment, recommended to enable, otherwise there will be many annoying warnings in the logs
    "enable_code_server": true, // Whether to enable the web code editor, if disabled, the Topology Configurator needs to be started in the background terminal
    "interface_name": "ens160", // Name of the main network card, the first IPv4 address of this network card will be used as the machine address, the master_address of slave nodes will be the machine address of the master node
    "debug": false, // Whether to enable debug mode (outputs debug logs)
    "instance_capacity": 400, // Capacity of container nodes on a single machine, the number of containers deployed on a single machine does not exceed this number
    "monitor_interval": 1, // Load monitoring sampling interval, in seconds
    "parallel_num": 64 // Degree of parallelism, the maximum number of container start/stop tasks executed simultaneously is parallel_num, the maximum number of link change tasks executed simultaneously is 4 * parallel_num 
  },
  "dependency": { // Configuration related to dependency services, keep default if unsure
    "etcd_port": 2379,
    "influxdb_port": 8086,
    "influxdb_token": "1145141919810",
    "influxdb_org": "satellite_emulator",
    "influxdb_bucket": "monitor",
    "code_server_port": 8079,
    "docker_host_path": "/var/run/docker.sock"
  },
  "device": { // Reserved functionality, can be ignored
    "FixPhysicalLink": [],
    "MultiplexPhysicalLink": []
  }
}
```

Specific command:
```bash
sudo ./NodeDaemon
```

If compiled from source, the `NodeDaemon` file is generally located in the `daemon/opensn-daemon/` directory.
If using the precompiled binary package, the `NodeDaemon` file is located in the `opensn-daemon/` directory of the extracted folder.

### 4.4 Expanding OpenSN Deployment Machines

Deploying slave nodes also requires the three steps of `kernel configuration`, `modifying configuration files`, and `starting the daemon`.

After each slave node successfully establishes a connection with the master node, the master node will output two new HTTP request logs.

## 5. Running Sample Programs and Expected Results

After OpenSN starts, you can access the OpenSN web console at `http://${master_address}:${listen_port}`.

In the `example-project` directory, we provide `example simulation configuration files` and `test topologies` of different network scales. Users can choose the test topology according to the hardware resources of their deployment.

| Minimum Hardware Specification | Corresponding Topology |
| :----------------------------: | :--------------------: |
|          2 cores 4GB           | example_6x11.json      |
|          8 cores 16GB          | example_31x31.json     |
|       8 cores 16GB * 2         | example_40x18.json     |
|      16 cores 32GB * 4         | example_72x22.json     |

1. Upload `example_emulation_configuration.json` and `example_MxN.json` to the console.
2. Observe the console log and wait for the topology upload to complete.
3. Click start.
4. Observe the console log and wait for the startup to complete.
5. Start the `Topology Configurator` in the `topology configurator directory` or the `console terminal`.
6. In the 3D earth view on the homepage, double-click on nodes or links to enter the details page. In the node details page, you can start a WebShell to operate inside the network node. In the link details page, you can start the link packet capture function to observe the data packets within the link.

## 6. Customizing Simulation Rules

Users can build rich simulation environments with OpenSN. OpenSN provides customization methods in the following three parts:

1. Container images
2. Topology controllers
3. Static configuration files

### 6.1 Customizing Node Types and Corresponding Behaviors

OpenSN supports any type of container image, as long as the image meets the following requirements:
1. The image has a configured entry point program and does not require specifying commands after startup, i.e., it can be run directly with the following command:
```bash
sudo docker run -d ${image_name_and_version}
```
2. The `/share` directory in the root directory of the image is not occupied, as this directory is used by the OpenSN daemon to mount configuration files.

Images that meet the above conditions can be used as node images in the OpenSN simulation environment.

Users can configure various network node roles such as routers, switches, user terminals, network services, etc., according to their simulation needs.

If the node needs to be dynamically configured during the simulation, the user needs to write a configuration program inside the container to dynamically configure the environment inside the container. The information needed for the configuration will be updated in real-time to the `/share/topo/${NODE_ID}` file, and the `${NODE_ID}` variable of each node is stored in the container's environment variables.

Specific implementation examples can refer to the [satellite node dynamic configuration program](container-base/satellite-router/daemon/) and the [ground station node dynamic configuration program](container-base/ground-station/daemon/).

### 6.2 Customizing Simulation Running Rules

Container images provide a customized method from the container's perspective, while `Topology Configurator` provides a macro view for customizing the entire simulation environment.

`Topology Configurator` defines the control logic during the simulation process, deciding the changes and triggering conditions of various network nodes and links in the simulation environment.

At the same time, this controller can also update the configuration

 files in the container's mount directory in real-time. Combined with the dynamic configurator in section 6.1, it achieves full-perspective control of the simulation environment.

The [Topology Configurator example](TopoConfigurators/Standard/main.py) is an example controller for a satellite network + ground station simulation environment. Users can refer to this example to implement customized controllers according to their simulation needs.

[OpenSN SDK For Python](TopoConfigurators/opensn/) is a Python-based OpenSN SDK. With this SDK, users can control objects such as nodes, links, geographic information, and container configurations in the OpenSN simulation environment using the Python language.

### 6.3 Customizing Simulation Scale

We also provide a customization method for different constellation scales and initial topology types under the same simulation rules, which are used in the configuration files in section 5.

Since the simulation scale may reach hundreds or thousands of satellites, users may need to use automation scripts to generate topology configuration files. You can refer to the [configuration file generation script](tools/namespace-generator/tle_generator.py) example we provided to customize your topology configuration files.

## 7. Known Issues

1. For large-scale network topologies, limited by the process scheduling capability of a single operating system, multiple machines are needed for simulation. Deploying on a single machine may reduce simulation stability and cause routing convergence issues.
2. In earlier versions, there may be container or link residue issues after the simulation ends. Although no such issues have occurred after subsequent fixes, the number of tests is insufficient. If there are container or link residue issues, you can run the [environment cleanup script](/daemon/scripts/delete_link_prefix.sh) to clean up the simulation environment.