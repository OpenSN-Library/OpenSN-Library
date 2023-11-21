#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <stdint.h>
#include <iproute2/bpf_elf.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define PACKET_HOST		0		/* To us		*/
#define PACKET_BROADCAST	1		/* To all		*/
#define PACKET_MULTICAST	2		/* To group		*/
#define PACKET_OTHERHOST	3		/* To someone else 	*/
#define PACKET_OUTGOING		4		/* Outgoing of any type */
#define PACKET_LOOPBACK		5		/* MC/BRD frame looped back */
#define PACKET_USER		6		/* To user space	*/

#define MAX_MAP_ENTRIES 256

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_MAP_ENTRIES);
	__type(key, __u32); // source IPv4 address
	__type(value, __u32); // redirect interface index 
} acc_map SEC(".maps");


SEC("eth1")
int tc_ingress(struct __sk_buff *skb)
{   
    if (skb != NULL) {
        bpf_printk("pkt_type is %d, ifindex is %u, ingress ifindex is %u",skb->pkt_type,skb->ifindex,skb->ingress_ifindex);
        bpf_skb_change_type(skb,PACKET_HOST);
    }
    
    return bpf_redirect_peer(82,0);
}

SEC("eth0")
int tc_egress(struct __sk_buff *skb)
{   
    
    if (skb != NULL) {
        bpf_printk("pkt_type is %d, ifindex is %u, ingress ifindex is %u",skb->pkt_type,skb->ifindex,skb->ingress_ifindex);
        
        bpf_skb_change_type(skb,PACKET_HOST);
    }
    return bpf_redirect_peer(84, 0);
}

char __license[] SEC("license") = "GPL";
