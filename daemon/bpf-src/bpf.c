#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <stdint.h>
#include <iproute2/bpf_elf.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ipv6.h>
#include <linux/icmpv6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define MAX_MAP_ENTRIES 256

struct hdr_cursor {
	void *pos;
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_MAP_ENTRIES);
	__type(key, __u64); // source Mac Address
	__type(value, __u32); // redirect interface index 
} egress_map SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 1);
	__type(key, __u64); // source Mac Address
	__type(value, __u64); // target interface mac address
} ingress_map SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(max_entries, 1);
	__type(key, __u32); // source Mac Address
	__type(value, __u32); // target interface mac address
} gateway_map SEC(".maps");

static __always_inline int parse_ethhdr(struct hdr_cursor *nh,
					void *data_end,
					struct ethhdr **ethhdr)
{
	struct ethhdr *eth = nh->pos;
	int hdrsize = sizeof(*eth);

	/* Byte-count bounds check; check if current pointer + size of header
	 * is after data_end.
	 */
	if (eth + 1 > data_end)
		return -1;

	nh->pos += hdrsize;
	*ethhdr = eth;
	
	return eth->h_proto; /* network-byte-order */
}

static __always_inline __u64 translate_mac(unsigned char* bytes) {
	__u64 res = 0;
	int i = 0;
	for (i=0; i<6; i++) {
		res <<= 8;
		res = (__u64)(bytes[i]) | res;
	}
	return res;
}

SEC("xdp")
int ingress(struct xdp_md *ctx)
{   
    void *data_end = (void *)(long)ctx->data_end;
	void *data = (void *)(long)ctx->data;
	struct ethhdr *eth;
	__u64 key = 0, value = 0;
	/* Default action XDP_PASS, imply everything we couldn't parse, or that
	 * we don't want to deal with, we just pass up the stack and let the
	 * kernel deal with it.
	 */
	__u32 action = XDP_PASS; /* Default action */

        /* These keep track of the next header type and iterator pointer */
	struct hdr_cursor nh;
	int nh_type;

	/* Start next header cursor position at data start */
	nh.pos = data;

	nh_type = parse_ethhdr(&nh, data_end, &eth);

	key = translate_mac(eth->h_source);
	void *value_ptr = bpf_map_lookup_elem(&ingress_map,&key);
	if (value_ptr == NULL) {
		return XDP_DROP;
	}
	
	value = (__u64*)
    // STEP1: lookup ingress map, get redirect ifindex
    // STEP2: redirect
}

SEC("xdp")
int egress(struct xdp_md *ctx)
{   
    
    // STEP1: parse packet, get src mac address, target mac address
    // STEP1: lookup ingress map, check redirect adress
    // STEP2: redirect/pass
}

char __license[] SEC("license") = "GPL";
