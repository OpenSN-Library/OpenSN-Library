

export class ResourceLimit {
    constructor() {
        this.nano_cpu = "";
        this.memory_byte = "";
    }

}

class NamespaceReqConfig {
    constructor() {
        this.image_map = {};
        this.container_envs = {};
        this.resource_map = {}
    }
}

export class CreateNamespaceReq {
    constructor() {
        this.name = "";
        this.ns_config = new NamespaceReqConfig();
        this.inst_config = [];
        this.link_config = [];
    }
}

export class NamespaceAbstrct {
    constructor() {
        this.name = "";
        this.instance_num = 0;
        this.link_num = 0;
        this.running = false;
        this.alloc_node_index = [];
    }
}

export class Namespace {
    constructor() {
        this.name = "";
        this.instance_num = 0;
        this.link_num = 0;
        this.running = false;
        this.alloc_node_index = [];
        this.instance_list = [];
        this.link_list = [];
    }
}