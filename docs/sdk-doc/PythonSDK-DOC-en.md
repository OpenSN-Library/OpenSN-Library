# OpenSN Python SDK Manual

## 1. Installation and setup

Requirements:

- Python 3.9 or later;
- network access to the platform's etcd service (default port 2379);
- the platform daemon (the background service that applies state) is running.

Install the SDK (its only dependency is an etcd client):

```bash
pip install -r TopoConfigurators/opensn/requirements.txt
# or install it as a local package
pip install -e TopoConfigurators/opensn
```

Provide the etcd address through environment variables, using either form:

```bash
export OPENSN_ETCD_ENDPOINTS="10.0.0.1:2379"
# or
export ETCD_ADDR="10.0.0.1"
export ETCD_PORT="2379"
```

You may list several addresses separated by commas. The current version uses the first one.

The first time you use the SDK in an environment, check the etcd client capabilities (transactions and watch):

```bash
cd TopoConfigurators
python3 opensn/tools/probe_etcd_api.py
```

If transactions are not available, set `OPENSN_ETCD_COMPAT=1` to fall back to one-by-one writes. This still works, but a batch write is no longer guaranteed to succeed as a whole.

When developing or upgrading the SDK, you can check that its fields match the platform code:

```bash
cd TopoConfigurators
python3 opensn/tools/check_contract.py
```

## 2. Connecting and closing

```python
from opensn import OpenSN

sn = OpenSN(["10.0.0.1:2379"])       # or use OpenSN.from_env()
sn.wait_ready(timeout=30, nodes=1)   # wait for etcd and at least one registered node

try:
    print([(n.node_index, n.ipv4, n.free_instance) for n in sn.nodes.list()])
finally:
    sn.close()
```

A context manager is also supported:

```python
from opensn import OpenSN

with OpenSN.from_env() as sn:
    sn.wait_ready()
    ...
```

Constructor parameters:

| Parameter | Default | Meaning |
| --- | --- | --- |
| `endpoints` | environment variables | etcd address list, for example `["10.0.0.1:2379"]` |
| `client` | `None` | pass an existing etcd client object, mainly for tests |
| `watch` | `True` | keep a local copy of etcd data updated by watching changes |
| `poll_interval` | `2.0` | refresh interval in seconds when watching is disabled |
| `chunk_ops` | `120` | maximum number of operations in one batch write |
| `cas_retries` | `3` | retries when a write loses a version conflict |
| `timeout` | `10.0` | etcd client timeout in seconds |
| `extra_prefixes` | `None` | extra key prefixes to keep in the local copy |

`wait_ready(timeout, nodes)` waits until etcd is reachable and the number of nodes under `/nodes` reaches `nodes`. It only proves that node records exist. It does not prove that the machines are alive.

## 3. Quick start

The example below defines an instance type, creates two instances and one link, then starts the emulation.

```python
from opensn import OpenSN
from opensn.schema.configs import InstanceSpec, LinkSpec

sn = OpenSN.from_env()
sn.wait_ready(timeout=30)

# 1) Define an instance type: the name, image and resources are your choice
sn.emulation.set_instance_types({
    "Router": {
        "image": "my-registry/router:1.0",
        "container_envs": {"MODE": "test"},
        "resource_limit": {"nano_cpu": "200M", "memory_byte": "256M"},
    }
})

# 2) Create two instances and one link (addresses are supplied by you)
topo = sn.emulation.deploy_topology(
    instances=[
        InstanceSpec("Router", extra={"role": "a"}),
        InstanceSpec("Router", extra={"role": "b"}),
    ],
    links=[
        LinkSpec(0, 1, address_infos=[{"IPV4": "10.0.0.1/30"}, {"IPV4": "10.0.0.2/30"}]),
    ],
)
print(topo.instance_ids)   # one ID per input item, in the same order
print(topo.link_ids)

# 3) Start the emulation (write the expected state; the daemon does the real start)
sn.emulation.start()
sn.emulation.wait_running(True, timeout=60)

# 4) Stop and clean up when done
sn.emulation.stop()
sn.emulation.reset(clean=True)
sn.close()
```

## 4. Platform objects

| Object | Meaning | How to access |
| --- | --- | --- |
| Node | a physical machine in the platform | `sn.nodes` |
| Instance | one container; it can act as a router, satellite, terminal or any other role | `sn.instances`, `sn.links.of(inst)` |
| Link | a network link between two instances | `sn.links` |
| Link parameter | delay, loss, bandwidth, connected or not | `link.set()`, `sn.links.parameter()` |
| Position | latitude, longitude and altitude of an instance | `sn.positions` |
| Instance config | content written to `/share/topo/<instance_id>` inside the container | `inst.push_config()` |
| Emulation config | mapping from instance type to image, resources and environment variables | `sn.emulation` |
| Terminal | a container shell or a capture shell opened on a node | `sn.shell` |
| Custom key | your own data outside the platform objects | `sn.kv` |

## 5. Reading data

At startup the SDK reads all platform data once, then keeps a local copy up to date by watching etcd. Normal reads come from that local copy, so they usually do not touch etcd and are fast.

### 5.1 Nodes

```python
nodes = sn.nodes.list()          # all nodes
node = sn.nodes.get(0)           # one node
mapping = sn.nodes.map()         # {node_index: Node}

print(node.ipv4, node.mac, node.free_instance, node.cpu_core)
```

Nodes are registered by the daemon. The SDK only reads them and provides no method to change a node record.

### 5.2 Instances

```python
instances = sn.instances.list()                 # all instances
node_instances = sn.instances.of_node(0)        # instances on one machine
routers = sn.instances.of_type("Router")        # instances of one type
running = sn.instances.started()                # instances whose start flag is true
inst = sn.instances.get("3f2a9c1d")             # by instance ID

print(inst.instance_id, inst.type, inst.node_index, inst.start)
print(inst.extra)                               # fields written by you
```

An instance ID is unique across the platform. If the same ID is found on more than one machine, `get` raises an error instead of picking one.

### 5.3 Links

```python
links = sn.links.list()                 # all links
link = sn.links.get("5c1e8a02")         # by link ID
linked = sn.links.of(inst)              # links connected to one instance
enabled = sn.links.active()             # links whose enable flag is true
between = sn.links.between(a, b)        # link between two instances, or None

print(link.addresses)                   # {"instance_id": "10.0.0.1/30", ...}
print(link.parameter)                   # parameter snapshot from link creation
print(link.get_parameters())            # current link parameters
```

A link between two machines is stored as two records. `sn.links.list()` merges them into one `Link` object, so you do not need to handle that yourself.

### 5.4 Positions

```python
position = sn.positions.get("3f2a9c1d")
all_positions = sn.positions.map()      # {instance_id: Position}

print(position.latitude, position.longitude, position.altitude)  # radians, meters
print(position.degrees())                                        # convert to degrees
```

Positions are used for display and calculation only. They do not affect the network directly.

### 5.5 Consistent snapshots and change notifications

```python
snap = sn.snapshot()                 # instances, links, positions and nodes at one moment
snap = sn.snapshot(force=True)       # read from etcd once before building the snapshot

print(len(snap.instances), len(snap.links))
print(snap.of_type("Router"))        # filter instances by type
print(snap.links_of(inst))           # links of one instance
```

Change notifications:

```python
for change in sn.watch(sn.links):
    print(change.kind, change.key, change.value)   # kind is "put" or "delete"
```

`watch` is a blocking iterator. Pass `sn.links`, `sn.instances`, `sn.positions` or `sn.nodes` to watch one kind of object, or pass nothing to watch everything.

## 6. Creating, changing and deleting

### 6.1 Deploying a topology in one call

```python
topo = sn.emulation.deploy_topology(
    instances=[InstanceSpec("Router"), InstanceSpec("Server")],
    links=[LinkSpec(0, 1)],
    nodes=None,              # optional: restrict the machines that may be used
    placement=None,          # optional: your own function that assigns machines
    charge_capacity=True,    # decrease the free capacity of the chosen machines
    dry_run=False,           # True: return what would be written, write nothing
)
```

`deploy_topology` does the following in order: checks that every type exists, assigns machines to instances, generates instance and link IDs, allocates link indexes, builds link records, writes instances and links, and updates the connections on both ends.

`InstanceSpec` fields:

| Field | Meaning |
| --- | --- |
| `type` | type name; it must already exist in the emulation configuration |
| `extra` | your own string key-value pairs, stored as is |
| `device_need` | device requirements, for example an exclusive network adapter |
| `node_index` | place the instance on this machine; leave empty to let the SDK choose |

`LinkSpec` fields:

| Field | Meaning |
| --- | --- |
| `a`, `b` | indexes of the two endpoints in the `instances` list |
| `type` | link type, default `"vlink"` |
| `init_parameter` | initial parameters when the link is created |
| `address_infos` | addresses of the two ends, always two elements |
| `extra` | your own key-value pairs |

To use your own placement rule, pass a function that receives the instance list and the node list and sets `node_index` on each instance:

```python
def place_on_nodes(instances, nodes):
    for index, inst in enumerate(instances):
        inst.node_index = nodes[index % len(nodes)].node_index

topo = sn.emulation.deploy_topology(instances, links, placement=place_on_nodes)
```

To see what would be written without writing, use `dry_run=True`. The returned `topo.plan` is a dictionary from key to content.

### 6.2 A single instance or link

```python
inst = sn.instances.create("Router", extra={"role": "edge"})
inst.remove()                       # remove the instance and its links, parameters, config and position

link = sn.links.create(
    a, b,
    type="vlink",
    address_infos=[{"IPV4": "10.0.0.5/30"}, {"IPV4": "10.0.0.6/30"}],
    init_parameter={"delay": 1000},
    extra={"note": "backup"},
)
link.remove()
sn.links.remove_between(a, b)       # remove all links between two instances; returns the count
```

When creating a link, the SDK allocates the link index, writes the link records and updates the connections on both instances. The default link type is `vlink`. If you use a type the platform does not know, the SDK warns you and the platform ignores the link.

### 6.3 Changing fields

For small changes, modify fields directly. The SDK reads the latest record, merges your changes and writes it back:

```python
inst.update(image="my-registry/router:2.0", extra={"role": "core"}, start=True)
link.update(enable=True, extra={"note": "updated"})
```

Editable instance fields: `name`, `image`, `extra`, `device_need`, `resource`, `start`.
Editable link fields: `enable`, `type`, `address_infos`, `extra`, `parameter`.

Instance ID, machine, link endpoints and link index are identity or topology fields and cannot be changed with `update`. To change endpoints, remove the link and create a new one.

`save()` writes every field of the object and may overwrite changes made by another program. Use it only when you know exactly what you are doing.

### 6.4 Deleting

```python
inst.delete_config()                # delete the instance config; the file in the container is removed too
sn.positions.delete(inst)           # delete one position
sn.positions.delete_all()           # delete all positions
link.remove_parameter()             # delete link parameters on both ends
link.remove_parameter(side=1)       # delete only one end
sn.emulation.remove_instance_type("DeprecatedType")   # delete an instance type
```

### 6.5 Emulation configuration (instance types)

```python
types = sn.emulation.instance_types()      # {type_name: InstanceTypeConfig}

sn.emulation.set_instance_types({          # replace all types (unlisted types are removed); running is preserved
    "Router": {"image": "router:1", "container_envs": {}, "resource_limit": {"nano_cpu": "100M", "memory_byte": "128M"}},
})

sn.emulation.update_instance_type("Router", image="router:2")   # change one type
sn.emulation.remove_instance_type("Router")                      # delete one type
```

`nano_cpu` accepts values such as `"50M"` and `"2G"` (powers of 1000). `memory_byte` accepts values such as `"128M"` and `"1G"` (powers of 1024).

### 6.6 Instance configuration (the file inside the container)

The platform writes the instance config, as is, to `/share/topo/<instance_id>` inside the container. The format is agreed between you and your container image. The SDK does not interpret it.

```python
config = {
    "name": inst.name,
    "links": {link.link_id: link.addresses for link in sn.links.of(inst)},
}
inst.push_config(config)                    # write the whole content
inst.push_config_if_absent(config)          # write only if it does not exist
inst.get_config()                           # read it back (JSON becomes an object)
inst.config_exists()                        # whether it exists
inst.delete_config()                        # delete it
```

Note: changing link connections does not update the instance config automatically. Your controller must call `push_config` explicitly.

### 6.7 Positions

```python
from opensn.schema.entities import Position

sn.positions.set(inst, Position(latitude=0.1, longitude=0.2, altitude=550000.0))
sn.positions.set_all({inst_a.instance_id: pos_a, inst_b.instance_id: pos_b})
sn.positions.delete(inst)
```

Latitude and longitude are in radians; altitude is in meters. You can also use `Position.from_degrees(latitude, longitude, altitude)`.

### 6.8 Link parameters

```python
link.set(delay_s=0.023, loss_ratio=0.015, bandwidth_bps=8_000_000, connect=True)
link.set(delay_s=0.023, side=0)          # write only the machine on one end
link.set_raw({"connect": 1, "delay": 12000}, side=1)   # provide the whole parameter table
print(link.get_parameters(side=0))       # read the current parameters of one end
link.remove_parameter(side=1)            # delete the parameters of one end
```

Parameter units:

| Parameter | How to write it | Meaning |
| --- | --- | --- |
| Delay | `delay_s=0.023` | seconds; stored internally in microseconds and increased by 1 microsecond when applied |
| Loss | `loss_ratio=0.015` | ratio; 0.015 means 1.5% |
| Bandwidth | `bandwidth_bps=8_000_000` | bits per second |
| Connected | `connect=True/False` | whether the link is connected |

The platform stores a complete parameter table. `set` reads the current table, applies your changes and keeps everything else; `set_raw` replaces the table with exactly what you pass. For a link between two machines, both ends get the same table by default; use `side=0/1` when the two ends must differ. Every value in a `set_raw` table must be an integer.

### 6.9 Start, stop and reset

```python
sn.emulation.start()        # set start=true on instances, enable=true on links, running=true
sn.emulation.stop()
sn.emulation.reset()        # delete links and instances, return capacity to the machines
sn.emulation.reset(clean=True)          # also remove parameters, instance configs and positions
sn.emulation.reset(keep_running=True)   # keep the running field, matching the old platform behavior

sn.emulation.status()                   # the current running field
sn.emulation.wait_running(True, timeout=60)
```

`start` and `stop` only write the expected state. The daemon performs the actual start and stop. `wait_running` waits for the `running` field in etcd; it does not mean that containers have finished starting.

### 6.10 Terminals (WebShell)

```python
shell = sn.shell.instance(inst)           # open a bash shell in the container by default
print(shell.url)                          # http://address:port/

shell = sn.shell.instance(inst, command="sh", args=["-c", "ip addr"])
shell = sn.shell.link(link)               # capture traffic on the node by default
shell = sn.shell.link(link, side=1, command="tcpdump", args=["-e", "-i", link.link_id])

shell = sn.shell.request(                 # fully custom command
    node_index=0,
    webshell_id="my_shell",
    command="bash",
    args=[],
    expire_minute=10,
)
```

The terminal process runs on the node machine, not inside the container; a container shell therefore uses `docker exec`. Requesting the same terminal ID again reuses the existing terminal.

### 6.11 Custom keys

Data outside the platform objects can live under your own keys:

```python
sn.kv.put("/my_state/plan", {"step": 1})          # stored as JSON
sn.kv.put("/my_state/raw", "any text")            # stored as text
print(sn.kv.get("/my_state/plan"))
print(sn.kv.get_raw("/my_state/plan"))            # raw bytes
print(sn.kv.get_prefix("/my_state/"))             # read a prefix
sn.kv.delete("/my_state/raw")
sn.kv.delete_prefix("/my_state/")

# Read, change and write back with a version check; conflicts are retried
sn.kv.cas_update("/my_state/plan", lambda cur: {"step": (cur or {}).get("step", 0) + 1})
value, revision = sn.kv.get_with_revision("/my_state/plan")
```

If you want custom keys to appear in the local copy, `snapshot()` and `watch()`, create the client with the prefixes:

```python
sn = OpenSN(extra_prefixes=["/my_state"])
```

## 7. Batch writes and the step loop

### 7.1 Batch writes

```python
with sn.batch():
    for link in sn.links.active():
        link.set(delay_s=0.02, connect=True)
    for inst in sn.instances.started():
        sn.positions.set(inst, position_of(inst))
```

Writes made inside the `with` block go into a queue and are sent when the block exits. Before sending, the SDK keeps only the last write for each key, skips writes whose value did not change, splits the rest into batches and retries on version conflicts.

### 7.2 Step loop

Controllers usually run at a fixed time step. The SDK provides a step loop:

```python
for step in sn.steps(interval=5):
    # this step's view comes from the local copy; no etcd reads
    for inst in step.instances.started():
        step.positions.set(inst, calculate_position(inst, step.now))

    for link in step.links.active():
        link.set(connect=True)

    # writes are sent before the next step begins
```

`step.now` is the start time of this step. `step.instances` and `step.links` are the object collections, and `step.positions` is the place to write positions. When the loop ends or raises, call `steps.close()` to send remaining writes, or use `with sn.steps(interval=5) as steps: for step in steps: ...`.

To write delay only when it changes by a certain amount (for example 1 millisecond), pass `delay_quantum_s=0.001`. This lowers precision and is off by default.

## 8. Units and data formats

| Item | How you write it | Value stored in etcd |
| --- | --- | --- |
| Delay | seconds (`0.023`) | integer microseconds (`23000`) |
| Loss | ratio (`0.015`) | integer in 1/10000 (`150`) |
| Bandwidth | bits per second (`8_000_000`) | bytes per second (`1000000`) |
| Connected | `True` / `False` | `1` / `0` |
| Position | radians, meters | radians, meters |
| Resource size | string (`"50M"`, `"128M"`) | integer |

- `extra` on an instance is a string-to-string dictionary. The platform does not interpret it; use it for role, index, latitude/longitude, orbit parameters and other business fields.
- The instance config can be any JSON, text or bytes.
- Link addresses use `address_infos`, always two elements. Each element is a string-to-string dictionary; `IPV4` is the common key.

## 9. Error handling

All SDK exceptions inherit from `OpenSNError`:

| Exception | Meaning |
| --- | --- |
| `NotFound` | the object or key does not exist |
| `ValidationError` | a pre-write check failed (field not editable, value out of range, and so on) |
| `Conflict` | a version conflict remained after retries |
| `NotReady` | etcd unreachable, wait timeout, or not enough nodes |
| `EtcdError` | etcd returned an error |
| `Unsupported` | the environment does not support it (for example a missing etcd client or transaction support) |

```python
import opensn

try:
    inst = sn.instances.get("not-exist")
except opensn.NotFound:
    print("instance does not exist")
except opensn.ValidationError as exc:
    print("pre-write check failed:", exc)
```

## 10. Limits and notes

1. **A write to etcd does not mean the change has taken effect.** The daemon applies it asynchronously. Check the daemon log, `docker ps` or `tc qdisc show`.
2. **etcd only.** If the deployment exposes only the daemon HTTP port and not etcd, the SDK cannot be used.
3. **No monitoring data.** Monitoring data is stored in InfluxDB, not in etcd.
4. **Node records are read-only.** The SDK has no method to change a node. Free capacity is maintained automatically when instances are created or removed.
5. **Reset behaves differently from the old HTTP API by default.** `reset()` sets `running` to false; pass `keep_running=True` for the old behavior.
6. **An instance cannot be moved to another machine.** Remove it and create a new one.
7. **A full save overwrites.** `save()` writes every field and may overwrite changes from other programs. Prefer `update()`.
8. **Unknown link types are ignored by the platform.** The SDK warns, but the platform implements only a limited set of link types.
9. **With several writers on one object**, version checks reduce overwrites, but a full save still follows the last writer.
10. **Run `probe_etcd_api.py` before first use** to confirm etcd client transactions and watch support.

## 11. Examples

### 11.1 Read and print an overview

```python
from opensn import OpenSN

with OpenSN.from_env() as sn:
    sn.wait_ready()
    print("nodes:")
    for node in sn.nodes.list():
        print(" ", node.node_index, node.ipv4, "free", node.free_instance)

    print("instances:")
    for inst in sn.instances.list():
        print(" ", inst.instance_id, inst.type, inst.node_index, inst.start)

    print("links:")
    for link in sn.links.list():
        print(" ", link.link_id, link.addresses, link.get_parameters())
```

### 11.2 Custom placement

```python
from opensn.schema.configs import InstanceSpec

def place_by_role(instances, nodes):
    """Put instances with role=master on node 0, spread the rest in turn."""
    others = [node for node in nodes if node.node_index != 0] or nodes
    cursor = 0
    for inst in instances:
        if inst.extra.get("role") == "master":
            inst.node_index = 0
        else:
            inst.node_index = others[cursor % len(others)].node_index
            cursor += 1

sn.emulation.deploy_topology(
    instances=[InstanceSpec("Router", extra={"role": "master"}),
               InstanceSpec("Router", extra={"role": "worker"}),
               InstanceSpec("Router", extra={"role": "worker"})],
    links=[],
    placement=place_by_role,
)
```

### 11.3 A simple controller loop

```python
from opensn import OpenSN

sn = OpenSN.from_env()
sn.wait_ready()
sn.emulation.wait_running(True, timeout=60)

# write an initial config for instances that do not have one
for inst in sn.instances.started():
    if not inst.config_exists():
        inst.push_config({"name": inst.name, "links": {}})

try:
    for step in sn.steps(interval=5):
        positions = {}
        for inst in step.instances.started():
            position = calculate_position(inst, step.now)   # implemented by you
            positions[inst.instance_id] = position
            step.positions.set(inst, position)

        for link in step.links.active():
            a, b = link.endpoints
            if a.instance_id not in positions or b.instance_id not in positions:
                continue
            link.set(
                delay_s=calculate_delay(positions[a.instance_id], positions[b.instance_id]),  # implemented by you
                connect=True,
                bandwidth_bps=8_000_000,
                loss_ratio=0.015,
            )
finally:
    sn.close()
```

### 11.4 Custom container config format

```python
def build_config(sn, inst):
    """Build config for your own image; the SDK only stores it as is."""
    links = sn.links.of(inst)
    return {
        "hostname": inst.instance_id,
        "interfaces": {
            link.link_id: link.addresses.get(inst.instance_id)
            for link in links
        },
    }

inst.push_config(build_config(sn, inst))
```

### 11.5 Track controller state in a custom key

```python
sn.kv.put("/my_controller/step", {"step": 0, "last_run": ""})

for step in sn.steps(interval=5):
    ...
    sn.kv.cas_update(
        "/my_controller/step",
        lambda cur: {"step": (cur or {}).get("step", 0) + 1, "last_run": str(step.now)},
    )
```

## Appendix A: API reference

### A.1 OpenSN

| Member | Meaning |
| --- | --- |
| `OpenSN(endpoints=None, client=None, watch=True, poll_interval=2.0, chunk_ops=120, cas_retries=3, timeout=10.0, extra_prefixes=None)` | create a client |
| `OpenSN.from_env(**kwargs)` | read the etcd address from environment variables |
| `nodes` / `instances` / `links` / `positions` / `emulation` / `shell` / `kv` | object collections |
| `wait_ready(timeout=30.0, nodes=1)` | wait for etcd and enough nodes |
| `snapshot(force=False)` | return the current snapshot |
| `watch(collection=None)` | blocking iterator over changes |
| `batch()` | start a batch write |
| `steps(interval=5.0, delay_quantum_s=None)` | create a step loop |
| `cache_size()` | bytes used by the local copy |
| `close()` | close the connection and background threads |
| `with OpenSN(...) as sn` | close automatically |

### A.2 Snapshot

| Member | Meaning |
| --- | --- |
| `instances` | list of instances |
| `links` | list of links |
| `positions` | `{instance_id: position}` |
| `nodes` | list of nodes |
| `of_type(type_name)` | filter instances by type |
| `links_of(inst)` | links of one instance |

### A.3 Nodes / Node

| Member | Meaning |
| --- | --- |
| `sn.nodes.list()` | all nodes |
| `sn.nodes.get(node_index)` | one node |
| `sn.nodes.map()` | `{node_index: Node}` |
| `node.node_index` | node index |
| `node.free_instance` | remaining instance capacity |
| `node.is_master_node` | whether this is the master node |
| `node.l_3_addr_v_4` / `node.l_3_addr_v_6` / `node.l_2_addr` | raw addresses (base64 strings) |
| `node.ipv4` / `node.ipv6` / `node.mac` | decoded addresses |
| `node.ns_instance_map` / `node.ns_link_map` | mappings maintained by the platform |
| `node.node_link_device_info` | device quota |
| `node.cpu_core` | CPU core count |

### A.4 Instances / Instance

| Member | Meaning |
| --- | --- |
| `sn.instances.list()` | all instances |
| `sn.instances.of_node(node_index)` | instances on one machine |
| `sn.instances.of_type(type_name)` | instances of one type |
| `sn.instances.started()` | instances whose `start` is true |
| `sn.instances.get(instance_id)` | get by ID |
| `sn.instances.create(type_name, extra=None, device_need=None, node_index=None, charge_capacity=True)` | create one instance |
| `inst.instance_id` / `inst.name` / `inst.type` / `inst.image` | basic fields |
| `inst.extra` | your own fields |
| `inst.device_need` | device requirements |
| `inst.resource` | resource limits |
| `inst.node_index` | machine |
| `inst.connections` | `{link_id: peer info}` |
| `inst.start` | whether it should run |
| `inst.start()` / `inst.stop()` | start/stop |
| `inst.remove()` | remove the instance and related data |
| `inst.update(**fields)` | change fields |
| `inst.save()` | save the whole record (careful) |
| `inst.push_config(content)` / `inst.push_config_if_absent(content)` | write instance config |
| `inst.get_config()` / `inst.config_exists()` / `inst.delete_config()` | read, check, delete instance config |
| `inst.connected_to(other)` / `inst.link_to(other)` | whether connected; link to the peer |

### A.5 Links / Link

| Member | Meaning |
| --- | --- |
| `sn.links.list()` | all links |
| `sn.links.get(link_id)` | get by ID |
| `sn.links.of(inst)` | links of one instance |
| `sn.links.active()` | links whose `enable` is true |
| `sn.links.between(a, b)` | link between two instances |
| `sn.links.create(a, b, type="vlink", address_infos=None, init_parameter=None, extra=None, enable=None)` | create a link |
| `sn.links.remove_between(a, b)` | remove links between two instances |
| `sn.links.parameter(link, side=None)` | read link parameters |
| `link.link_id` / `link.type` / `link.enable` | basic fields |
| `link.end_infos` | endpoint info (`EndInfo` list, exactly two) |
| `link.address_infos` | endpoint addresses (exactly two) |
| `link.link_index` / `link.cross_machine` / `link.node_index` / `link.extra` | platform fields |
| `link.parameter` | parameter snapshot from link creation |
| `link.endpoints` | `(instance, instance)` |
| `link.addresses` | `{instance_id: address}` |
| `link.save()` | save the whole record |
| `link.update(**fields)` | change fields |
| `link.connect(side=None)` / `link.disconnect(side=None)` | connect/disconnect |
| `link.remove()` | remove the link |
| `link.set(delay_s=None, loss_ratio=None, bandwidth_bps=None, connect=None, side=None)` | change parameters |
| `link.set_raw(parameter, side=None)` | replace the whole parameter table |
| `link.get_parameters(side=None)` | read parameters |
| `link.remove_parameter(side=None)` | delete parameters |

### A.6 Positions / Position

| Member | Meaning |
| --- | --- |
| `sn.positions.get(instance_id)` | read a position |
| `sn.positions.map()` | all positions |
| `sn.positions.set(inst, position)` | write a position |
| `sn.positions.set_all(mapping)` | write several positions |
| `sn.positions.delete(inst)` | delete a position |
| `sn.positions.delete_all()` | delete all positions |
| `Position(latitude=0.0, longitude=0.0, altitude=0.0)` | create a position |
| `Position.from_degrees(latitude, longitude, altitude=0.0)` | create from degrees |
| `position.degrees()` | latitude, longitude and altitude in degrees |

### A.7 Emulation / InstanceTypeConfig / TopologyResult

| Member | Meaning |
| --- | --- |
| `sn.emulation.instance_types()` | read all instance types |
| `sn.emulation.set_instance_types(mapping)` | replace all instance types |
| `sn.emulation.update_instance_type(name, image=None, container_envs=None, resource_limit=None)` | change one type |
| `sn.emulation.remove_instance_type(name)` | remove one type |
| `sn.emulation.status()` | read `running` |
| `sn.emulation.wait_running(expect=True, timeout=60)` | wait for `running` to reach a value |
| `sn.emulation.deploy_topology(instances, links=None, nodes=None, placement=None, charge_capacity=True, dry_run=False)` | create a topology |
| `sn.emulation.start()` / `stop()` | start/stop the emulation |
| `sn.emulation.reset(clean=False, keep_running=False)` | reset the emulation |
| `InstanceTypeConfig.image` / `.container_envs` / `.resource_limit` | type fields |
| `TopologyResult.instances` / `.links` / `.plan` | creation result |
| `TopologyResult.instance_ids` / `.link_ids` | IDs matching the input order |

### A.8 Shell / ShellInfo

| Member | Meaning |
| --- | --- |
| `sn.shell.request(node_index, webshell_id, command, args, expire_minute=None)` | custom terminal |
| `sn.shell.instance(inst, expire_minute=None, command="docker", args=None)` | container terminal |
| `sn.shell.link(link, expire_minute=None, side=0, command="tcpdump", args=None)` | link terminal |
| `ShellInfo.webshell_id` / `.addr` / `.port` / `.pid` | terminal info |
| `ShellInfo.url` | terminal URL |

### A.9 KV

| Member | Meaning |
| --- | --- |
| `sn.kv.get(key, default=None)` | read and parse JSON |
| `sn.kv.get_raw(key)` | read raw bytes |
| `sn.kv.get_prefix(prefix)` | read a prefix |
| `sn.kv.require(key)` | read, raise if missing |
| `sn.kv.get_with_revision(key)` | read value and revision |
| `sn.kv.cas_update(key, update_fn, tries=5)` | read-modify-write with a version check |
| `sn.kv.put(key, value)` | write |
| `sn.kv.put_if_absent(key, value)` | write only if absent |
| `sn.kv.delete(key)` | delete |
| `sn.kv.delete_prefix(prefix)` | delete a prefix |

### A.10 Steps / Step / StepPositions

| Member | Meaning |
| --- | --- |
| `sn.steps(interval=5.0, delay_quantum_s=None)` | create a step loop; iterable and usable as a context manager |
| `steps.close()` | send remaining writes and finish |
| `step.now` | start time of this step |
| `step.interval` | step length in seconds |
| `step.instances` / `step.links` | object collections for this step |
| `step.positions.set(inst, position)` / `step.positions.set_all(mapping)` | write positions in this step |
| `step.flush()` | send this step's writes immediately |

### A.11 WriteBatch

| Member | Meaning |
| --- | --- |
| `with sn.batch() as batch` | start a batch write |
| `batch.put(key, obj, expect=None)` | write an object or dictionary |
| `batch.put_bytes(key, value, expect=None)` | write bytes |
| `batch.delete(key)` | delete a key |
| `batch.size()` | number of queued items |
| `batch.operation_count()` | estimated number of operations |
| `batch.commit()` | send and return the number of operations |

`expect` sets a version condition: omit it for a plain write; pass `-1` to require that the key does not exist; pass a revision to require that exact revision.

### A.12 Change

| Member | Meaning |
| --- | --- |
| `change.kind` | `"put"` or `"delete"` |
| `change.key` | the key that changed |
| `change.value` | the new value (parsed JSON), `None` for a delete |

### A.13 Exceptions

| Class | Meaning |
| --- | --- |
| `OpenSNError` | base class of all SDK exceptions |
| `NotFound` | the object or key does not exist |
| `ValidationError` | a pre-write check failed |
| `Conflict` | a version conflict remained after retries |
| `NotReady` | etcd unavailable, wait timeout, or not enough nodes |
| `EtcdError` | etcd returned an error |
| `Unsupported` | the environment does not support this capability |

## Appendix B: Data object fields

| Object | Fields |
| --- | --- |
| `InstanceSpec` | `type`, `extra`, `device_need`, `node_index` |
| `LinkSpec` | `a`, `b`, `type`, `init_parameter`, `address_infos`, `extra` |
| `EndInfo` | `instance_id`, `instance_type`, `end_node_index` |
| `Connection` | `link_id`, `instance_id`, `instance_type`, `end_node_index` |
| `DeviceRequire` | `dev_name`, `need_num`, `is_mutex` |
| `ResourceLimit` | `nano_cpu`, `memory_byte` |
| `EmulationConfig` | `running`, `type_config` |
| `InstanceTypeConfig` | `image`, `container_envs`, `resource_limit` |

## Appendix C: Tools, environment variables and helper functions

### C.1 Command-line tools

| Tool | Purpose |
| --- | --- |
| `python3 opensn/tools/probe_etcd_api.py` | check etcd client transactions and watch support |
| `python3 opensn/tools/check_contract.py` | compare SDK fields and keys with the platform code |
| `make test` (under TopoConfigurators) | run SDK unit tests and the example smoke test |
| `make check` | run the contract check |

### C.2 Environment variables

| Variable | Meaning |
| --- | --- |
| `OPENSN_ETCD_ENDPOINTS` | etcd addresses, separated by commas |
| `ETCD_ADDR` / `ETCD_PORT` | etcd address and port used when the variable above is absent |
| `OPENSN_ETCD_COMPAT` | set to `1` to fall back to one-by-one writes when transactions are unavailable |

### C.3 Unit conversion functions

```python
from opensn.schema.units import (
    parse_size_decimal, parse_size_binary,
    delay_us, loss_raw, bandwidth_bytes,
    rad_from_deg, deg_from_rad,
)
```

| Function | Purpose |
| --- | --- |
| `parse_size_decimal("50M")` | parse a size using powers of 1000 |
| `parse_size_binary("128M")` | parse a size using powers of 1024 |
| `delay_us(0.023)` | seconds to microseconds |
| `loss_raw(0.015)` | ratio to 1/10000 units |
| `bandwidth_bytes(8_000_000)` | bits per second to bytes per second |
| `rad_from_deg(deg)` / `deg_from_rad(rad)` | convert between degrees and radians |
