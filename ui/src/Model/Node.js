
export class NodeAbstract {
    constructor() {
        this.node_id = 0;
        this.free_instance = 0;
        this.is_master_node = "";
        this.l_3_addr_v_4 = "";
        this.l_3_addr_v_6 = "";
        this.l_2_addr = "";
    }
}

export class Node {
    constructor() {
        this.node_id = 0;
        this.free_instance = 0;
        this.is_master_node = "";
        this.l_3_addr_v_4 = "";
        this.l_3_addr_v_6 = "";
        this.l_2_addr = "";
        this.ns_instance_map = {}
        this.ns_link_map = {}
    }
}