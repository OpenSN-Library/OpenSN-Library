

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

